// Copyright (C) 2024 The go-matter Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matter

import (
	"context"
	"fmt"
	"net"
	"runtime"

	"github.com/cybergarage/go-logger/log"
	"github.com/YashubuStudio/go-matter-pack/matter/ble"
	"github.com/YashubuStudio/go-matter-pack/matter/errors"
	"github.com/YashubuStudio/go-matter-pack/matter/mdns"
)

// commissioner represents a commissioner.
type commissioner struct {
	ble.Central
	discoverer mdns.Discoverer
	enableBLE  bool
	enableMDNS bool
}

// CommissionerOption represents a configuration option for a commissioner.
type CommissionerOption func(*commissionerOptions)

type commissionerOptions struct {
	enableBLE  bool
	enableMDNS bool
}

// WithCommissionerBLEEnabled toggles BLE scanning.
func WithCommissionerBLEEnabled(enabled bool) CommissionerOption {
	return func(opts *commissionerOptions) {
		opts.enableBLE = enabled
	}
}

// WithCommissionerMDNSEnabled toggles mDNS discovery.
func WithCommissionerMDNSEnabled(enabled bool) CommissionerOption {
	return func(opts *commissionerOptions) {
		opts.enableMDNS = enabled
	}
}

// NewCommissioner returns a new commissioner.
func NewCommissioner() Commissioner {
	return NewCommissionerWithOptions()
}

// NewCommissionerWithOptions returns a new commissioner configured with options.
func NewCommissionerWithOptions(options ...CommissionerOption) Commissioner {
	opts := commissionerOptions{}
	for _, apply := range options {
		apply(&opts)
	}

	if runtime.GOOS == "windows" && opts.enableBLE {
		log.Warn("BLE commissioning is disabled on Windows; ignoring enable flag.")
		opts.enableBLE = false
	}

	com := &commissioner{
		Central:    ble.NewCentral(),
		discoverer: mdns.NewDiscoverer(),
		enableBLE:  opts.enableBLE,
		enableMDNS: opts.enableMDNS,
	}
	return com
}

// Scannar returns the BLE scanner.
func (cmr *commissioner) Scannar() ble.Scanner {
	return cmr.Central
}

// Discoverer returns the mDNS discoverer.
func (cmr *commissioner) Discoverer() mdns.Discoverer {
	return cmr.discoverer
}

// Discover discovers commissionable devices.
// 5.4.3. Discovery by Commissioner.
func (cmr *commissioner) Discover(ctx context.Context, query Query) ([]CommissionableDevice, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultDiscoveryTimeout)
		defer cancel()
	}

	if addr, port, ok := onNetworkAddress(query); ok {
		payload, ok := query.OnboardingPayload()
		if !ok {
			return nil, fmt.Errorf("%w: onboarding payload required for on-network commissioning", errors.ErrInvalid)
		}
		if port <= 0 {
			port = mdns.Port
		}
		return []CommissionableDevice{newOnNetworkDevice(addr, port, payload)}, nil
	}

	scanNodes := func(ctx context.Context) ([]CommissionableDevice, error) {
		var devs []CommissionableDevice
		scanner := cmr.Scannar()
		err := scanner.Scan(ctx)
		if err != nil {
			return nil, err
		}
		for _, bleDev := range scanner.DiscoveredDevices() {
			if !bleDev.IsCommissionable() {
				continue
			}
			bleService, err := bleDev.Service()
			if err != nil {
				continue
			}
			devs = append(devs, newBLEDevice(bleDev, bleService))
		}
		return devs, nil
	}

	discoverNodes := func(ctx context.Context) ([]CommissionableDevice, error) {
		var devs []CommissionableDevice
		dnsQueryOpts := []mdns.QueryOption{
			mdns.WithQueryService(mdns.CommissionableNodeService),
		}
		if payload, ok := query.OnboardingPayload(); ok {
			dnsQueryOpts = append(dnsQueryOpts,
				mdns.WithQueryOnboardingPayload(payload),
			)
		}
		nodes, err := cmr.discoverer.Search(ctx, mdns.NewQuery(dnsQueryOpts...))
		if err != nil {
			return nil, err
		}
		for _, entry := range nodes {
			devs = append(devs, newMDNSDevice(entry))
		}
		return devs, nil
	}

	// Run BLE scan and mDNS discovery in parallel
	type result struct {
		devs []CommissionableDevice
		err  error
	}

	var discoverers []func(context.Context) ([]CommissionableDevice, error)
	if cmr.enableBLE {
		discoverers = append(discoverers, scanNodes)
	}
	if cmr.enableMDNS {
		discoverers = append(discoverers, discoverNodes)
	}
	if len(discoverers) == 0 {
		return nil, fmt.Errorf("%w: discovery is disabled (enable BLE or mDNS)", ErrDisabled)
	}

	// Use a single channel to collect results symmetrically
	done := make(chan result, len(discoverers))

	for _, discover := range discoverers {
		go func(fn func(context.Context) ([]CommissionableDevice, error)) {
			d, e := fn(ctx)
			done <- result{devs: d, err: e}
		}(discover)
	}

	var devs []CommissionableDevice

	// Collect results; treat timeouts as normal (skip)
	for range discoverers {
		r := <-done
		if r.err != nil && !errors.Is(r.err, context.DeadlineExceeded) {
			return nil, r.err
		}
		devs = append(devs, r.devs...)
	}

	return devs, nil
}

// 5.5. Commissioning Flows.
func (cmr *commissioner) Commission(ctx context.Context, payload OnboardingPayload) (Commissionee, error) {
	query := NewQuery(
		WithQueryOnboardingPayload(payload),
	)
	return cmr.commissionWithQuery(ctx, payload, query)
}

// CommissionOnNetwork commissions a device with a direct on-network address.
func (cmr *commissioner) CommissionOnNetwork(ctx context.Context, payload OnboardingPayload, address net.IP, port int) (Commissionee, error) {
	query := NewQuery(
		WithQueryOnboardingPayload(payload),
		WithQueryOnNetworkAddress(address, port),
	)
	return cmr.commissionWithQuery(ctx, payload, query)
}

func (cmr *commissioner) commissionWithQuery(ctx context.Context, payload OnboardingPayload, query Query) (Commissionee, error) {
	devs, err := cmr.Discover(ctx, query)
	if err != nil {
		return nil, err
	}

	log.Infof("Discovered device: %d", len(devs))
	for n, dev := range devs {
		log.Infof("[%d] %s", n, dev.String())
	}

	isCommissionableDevicePayload := func(dev CommissionableDevice, payload OnboardingPayload) bool {
		return dev.VendorID().Equal(VendorID(payload.VendorID())) &&
			dev.ProductID().Equal(ProductID(payload.ProductID())) &&
			dev.Discriminator().Equal(Discriminator(payload.Discriminator()))
	}

	for _, dev := range devs {
		if isCommissionableDevicePayload(dev, payload) {
			err = dev.Commission(ctx, payload)
			if err != nil {
				return nil, fmt.Errorf("%w to commission device (%s): %w", ErrFailed, dev.String(), err)
			}
			return newCommissioneeWithDevice(dev), nil
		}
	}

	return nil, fmt.Errorf("%w: no matching commissionable device found (payload=%s)", ErrNotFound, payload.String())
}

// Start starts the commissioner.
func (cmr *commissioner) Start() error {
	if !cmr.enableMDNS {
		return nil
	}
	err := cmr.discoverer.Start()
	if err != nil {
		return err
	}
	return nil
}

// Stop stops the commissioner.
func (cmr *commissioner) Stop() error {
	if !cmr.enableMDNS {
		return nil
	}
	err := cmr.discoverer.Stop()
	if err != nil {
		return err
	}
	return nil
}

func onNetworkAddress(query Query) (net.IP, int, bool) {
	type onNetworkQuery interface {
		OnNetworkAddress() (net.IP, int, bool)
	}
	q, ok := query.(onNetworkQuery)
	if !ok {
		return nil, 0, false
	}
	return q.OnNetworkAddress()
}
