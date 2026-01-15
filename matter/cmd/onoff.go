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
	"path/filepath"
	"strings"
	"time"

	"github.com/YashubuStudio/go-matter-pack/internal/app"
	"github.com/YashubuStudio/go-matter-pack/internal/matterctrl"
	"github.com/YashubuStudio/go-matter-pack/internal/store"
	"github.com/YashubuStudio/go-matter-pack/internal/usecase"
	"github.com/spf13/cobra"
)

const defaultRegistryFilename = "registry.json"

func init() {
	onoffCmd.AddCommand(onoffOnCmd)
	onoffCmd.AddCommand(onoffOffCmd)
	onoffCmd.AddCommand(onoffToggleCmd)
	onoffCmd.AddCommand(onoffStatusCmd)
	rootCmd.AddCommand(onoffCmd)

	onoffCmd.PersistentFlags().String("state-dir", "", "state directory (defaults to XDG state home)")
	onoffCmd.PersistentFlags().Duration("timeout", 5*time.Second, "command timeout")
}

var onoffCmd = &cobra.Command{ // nolint:exhaustruct
	Use:   "onoff",
	Short: "Control bridged On/Off devices.",
	Long:  "Control bridged On/Off devices.",
}

var onoffOnCmd = &cobra.Command{ // nolint:exhaustruct
	Use:   "on <unique id>",
	Short: "Turn on a device by unique ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service, ctx, cancel, err := newOnOffService(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		return service.On(ctx, args[0])
	},
}

var onoffOffCmd = &cobra.Command{ // nolint:exhaustruct
	Use:   "off <unique id>",
	Short: "Turn off a device by unique ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service, ctx, cancel, err := newOnOffService(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		return service.Off(ctx, args[0])
	},
}

var onoffToggleCmd = &cobra.Command{ // nolint:exhaustruct
	Use:   "toggle <unique id>",
	Short: "Toggle a device by unique ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service, ctx, cancel, err := newOnOffService(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		return service.Toggle(ctx, args[0])
	},
}

var onoffStatusCmd = &cobra.Command{ // nolint:exhaustruct
	Use:   "status <unique id>",
	Short: "Read On/Off state by unique ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service, ctx, cancel, err := newOnOffService(cmd)
		if err != nil {
			return err
		}
		defer cancel()
		state, err := service.State(ctx, args[0])
		if err != nil {
			return err
		}
		if state {
			outputf("on\n")
		} else {
			outputf("off\n")
		}
		return nil
	},
}

func newOnOffService(cmd *cobra.Command) (*usecase.OnOffService, context.Context, context.CancelFunc, error) {
	stateDir, err := cmd.Flags().GetString("state-dir")
	if err != nil {
		return nil, nil, nil, err
	}
	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return nil, nil, nil, err
	}
	if stateDir == "" {
		stateDir = app.StateDir(defaultAppName)
	}
	stateDir = strings.TrimSpace(stateDir)
	registryPath := filepath.Join(stateDir, defaultRegistryFilename)
	stateStore := store.NewJSONFileStore(registryPath)

	ctrl := matterctrl.NewNoopController()
	if ctrl == nil {
		return nil, nil, nil, fmt.Errorf("failed to create controller")
	}
	service := usecase.NewOnOffService(ctrl, stateStore)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return service, ctx, cancel, nil
}
