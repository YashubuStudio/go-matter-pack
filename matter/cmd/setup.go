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

package cmd

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cybergarage/go-logger/log"
	"github.com/YashubuStudio/go-matter-pack/internal/app"
	"github.com/YashubuStudio/go-matter-pack/internal/commission"
	"github.com/YashubuStudio/go-matter-pack/internal/store"
	"github.com/YashubuStudio/go-matter-pack/internal/usecase"
	"github.com/YashubuStudio/go-matter-pack/matter"
	"github.com/YashubuStudio/go-matter-pack/matter/mdns"
	"github.com/spf13/cobra"
)

const defaultAppName = "go-matter-pack"

func init() {
	setupCmd.AddCommand(setupCommissionCmd)
	rootCmd.AddCommand(setupCmd)

	setupCommissionCmd.Flags().String("qr", "", "QR onboarding payload")
	setupCommissionCmd.Flags().String("code", "", "manual pairing code")
	setupCommissionCmd.Flags().Uint64("node-id", 0, "target node ID")
	setupCommissionCmd.Flags().String("state-dir", "", "state directory (defaults to XDG state home)")
	setupCommissionCmd.Flags().Duration("timeout", 30*time.Second, "commissioning timeout")
	setupCommissionCmd.Flags().Bool("import-only", false, "only store onboarding payload without commissioning")
	setupCommissionCmd.Flags().String("address", "", "on-network device address (ip or ip:port)")
}

var setupCmd = &cobra.Command{ // nolint:exhaustruct
	Use:   "setup",
	Short: "Setup commands for go-matter-pack.",
	Long:  "Setup commands for go-matter-pack.",
}

var setupCommissionCmd = &cobra.Command{ // nolint:exhaustruct
	Use:   "commission",
	Short: "Commission a Matter bridge and store onboarding data.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		qrPayload, err := cmd.Flags().GetString("qr")
		if err != nil {
			return err
		}
		codePayload, err := cmd.Flags().GetString("code")
		if err != nil {
			return err
		}
		if (qrPayload == "" && codePayload == "") || (qrPayload != "" && codePayload != "") {
			return fmt.Errorf("specify exactly one of --qr or --code")
		}
		nodeID, err := cmd.Flags().GetUint64("node-id")
		if err != nil {
			return err
		}
		stateDir, err := cmd.Flags().GetString("state-dir")
		if err != nil {
			return err
		}
		timeout, err := cmd.Flags().GetDuration("timeout")
		if err != nil {
			return err
		}
		importOnly, err := cmd.Flags().GetBool("import-only")
		if err != nil {
			return err
		}
		address, err := cmd.Flags().GetString("address")
		if err != nil {
			return err
		}

		payload := codePayload
		if qrPayload != "" {
			payload = qrPayload
		}
		if stateDir == "" {
			stateDir = app.StateDir(defaultAppName)
		}
		statePath := filepath.Join(stateDir, "commission.json")
		stateStore := store.NewJSONFileStore(statePath)

		service := usecase.NewCommissionService(SharedCommissioner(), stateStore)
		if importOnly {
			if _, err := service.ImportPayload(context.Background(), nodeID, payload); err != nil {
				return err
			}
			log.Infof("Imported commissioning payload to %s", statePath)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		ip, port, err := parseOnNetworkAddress(address)
		if err != nil {
			return err
		}

		var state commission.State
		var commissionee matter.Commissionee
		if ip != nil {
			state, commissionee, err = service.CommissionOnNetwork(ctx, nodeID, payload, ip, port)
		} else {
			state, commissionee, err = service.Commission(ctx, nodeID, payload)
		}
		if err != nil {
			return err
		}
		log.Infof("Successfully commissioned device: %s", commissionee.String())
		if state.Result != nil {
			log.Infof("Saved commissioning result to %s", statePath)
		}
		return nil
	},
}

func parseOnNetworkAddress(address string) (net.IP, int, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, 0, nil
	}
	if ip := net.ParseIP(address); ip != nil {
		return ip, mdns.Port, nil
	}
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, 0, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, 0, fmt.Errorf("invalid IP address: %s", host)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return nil, 0, fmt.Errorf("invalid port: %s", portStr)
	}
	return ip, port, nil
}
