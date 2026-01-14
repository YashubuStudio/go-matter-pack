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
	"errors"
	"fmt"
	"strings"

	"github.com/cybergarage/go-logger/log"
	"github.com/cybergarage/go-matter/matter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	VerboseParamStr     = "verbose"
	DebugParamStr       = "debug"
	ConfigParamStr      = "config"
	EnableBLEParamStr   = "enable-ble"
	EnableMDNSParamStr  = "enable-mdns"
	commissionerStarted = false
)

var rootCmd = &cobra.Command{ // nolint:exhaustruct
	Use:               "matterctl",
	Version:           matter.Version,
	Short:             "",
	Long:              "",
	DisableAutoGenTag: true,
}

func GetRootCommand() *cobra.Command {
	return rootCmd
}

var sharedCommissioner matter.Commissioner

func SharedCommissioner() matter.Commissioner {
	return sharedCommissioner
}

func Execute() error {
	err := rootCmd.Execute()
	if sharedCommissioner == nil {
		return err
	}
	return errors.Join(err, sharedCommissioner.Stop())
}

func loadConfig() error {
	configFile := viper.GetString(ConfigParamStr)
	if configFile == "" {
		return nil
	}
	viper.SetConfigFile(configFile)
	return viper.ReadInConfig()
}

func configureLogging() {
	verbose := viper.GetBool(VerboseParamStr)
	debug := viper.GetBool(DebugParamStr)
	if debug {
		verbose = true
	}
	if verbose {
		enableStdoutVerbose(verbose, debug)
	}
}

func initCommissioner() error {
	if commissionerStarted {
		return nil
	}

	if err := loadConfig(); err != nil {
		return err
	}
	configureLogging()

	enableBLE := viper.GetBool(EnableBLEParamStr)
	enableMDNS := viper.GetBool(EnableMDNSParamStr)
	sharedCommissioner = matter.NewCommissionerWithOptions(
		matter.WithCommissionerBLEEnabled(enableBLE),
		matter.WithCommissionerMDNSEnabled(enableMDNS),
	)

	if err := sharedCommissioner.Start(); err != nil {
		return err
	}
	commissionerStarted = true
	return nil
}

func init() {
	viper.SetEnvPrefix("matter_ctl")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := initCommissioner(); err != nil {
			log.Error(err)
			return err
		}
		return nil
	}

	viper.SetDefault(FormatParamStr, FormatTableStr)
	rootCmd.PersistentFlags().String(FormatParamStr, FormatTableStr, fmt.Sprintf("output format: %s", strings.Join(allSupportedFormats(), "|")))
	viper.BindPFlag(FormatParamStr, rootCmd.PersistentFlags().Lookup(FormatParamStr))
	viper.BindEnv(FormatParamStr) // MATTER_CTL_FORMAT

	viper.SetDefault(ConfigParamStr, "")
	rootCmd.PersistentFlags().String(ConfigParamStr, "", "config file path")
	viper.BindPFlag(ConfigParamStr, rootCmd.PersistentFlags().Lookup(ConfigParamStr))
	viper.BindEnv(ConfigParamStr) // MATTER_CTL_CONFIG

	viper.SetDefault(VerboseParamStr, false)
	rootCmd.PersistentFlags().Bool((VerboseParamStr), false, "enable verbose output")
	viper.BindPFlag(VerboseParamStr, rootCmd.PersistentFlags().Lookup(VerboseParamStr))
	viper.BindEnv(VerboseParamStr) // MATTER_CTL_VERBOSE

	viper.SetDefault(DebugParamStr, false)
	rootCmd.PersistentFlags().Bool((DebugParamStr), false, "enable debug output")
	viper.BindPFlag(DebugParamStr, rootCmd.PersistentFlags().Lookup(DebugParamStr))
	viper.BindEnv(DebugParamStr) // MATTER_CTL_DEBUG

	viper.SetDefault(EnableBLEParamStr, false)
	rootCmd.PersistentFlags().Bool(EnableBLEParamStr, false, "enable BLE commissioning")
	viper.BindPFlag(EnableBLEParamStr, rootCmd.PersistentFlags().Lookup(EnableBLEParamStr))
	viper.BindEnv(EnableBLEParamStr) // MATTER_CTL_ENABLE_BLE

	viper.SetDefault(EnableMDNSParamStr, false)
	rootCmd.PersistentFlags().Bool(EnableMDNSParamStr, false, "enable mDNS discovery")
	viper.BindPFlag(EnableMDNSParamStr, rootCmd.PersistentFlags().Lookup(EnableMDNSParamStr))
	viper.BindEnv(EnableMDNSParamStr) // MATTER_CTL_ENABLE_MDNS
}
