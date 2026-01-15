// Copyright (C) 2025 The go-matter Authors. All rights reserved.
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
)

type onNetworkDevice struct {
	*baseDevice
	address net.IP
	port    int
	payload OnboardingPayload
}

func newOnNetworkDevice(address net.IP, port int, payload OnboardingPayload) CommissionableDevice {
	return &onNetworkDevice{
		baseDevice: &baseDevice{},
		address:    address,
		port:       port,
		payload:    payload,
	}
}

// VendorID represents a vendor ID.
// 2.5.2. Vendor Identifier (Vendor ID, VID).
func (dev *onNetworkDevice) VendorID() VendorID {
	return VendorID(dev.payload.VendorID())
}

// ProductID represents a product ID.
// 2.5.3. Product Identifier (Product ID, PID).
func (dev *onNetworkDevice) ProductID() ProductID {
	return ProductID(dev.payload.ProductID())
}

// Discriminator represents a discriminator.
// 2.5.6. Discriminator.
func (dev *onNetworkDevice) Discriminator() Discriminator {
	return Discriminator(dev.payload.Discriminator())
}

// Commission commissions the node with the given commissioning options.
func (dev *onNetworkDevice) Commission(_ context.Context, _ OnboardingPayload) error {
	return nil
}

// String returns the string representation of the on-network device.
func (dev *onNetworkDevice) String() string {
	return fmt.Sprintf("%s, Address: %s:%d", dev.baseDevice.String(dev), dev.address.String(), dev.port)
}

// MarshalObject returns an object suitable for marshaling to JSON.
func (dev *onNetworkDevice) MarshalObject() any {
	return struct {
		Discriminator uint16 `json:"discriminator"`
		VendorID      uint16 `json:"vendorId"`
		ProductID     uint16 `json:"productId"`
		Address       string `json:"address"`
		Port          int    `json:"port"`
	}{
		Discriminator: uint16(dev.Discriminator()),
		VendorID:      uint16(dev.VendorID()),
		ProductID:     uint16(dev.ProductID()),
		Address:       dev.address.String(),
		Port:          dev.port,
	}
}
