/*
Copyright © 2026 Gumlet Pte. Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gumlet/pkg/ui"
)

var noColor bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gumlet",
	Short: "Gumlet CLI — manage video and image assets from your terminal",
	Long: `Gumlet CLI puts Gumlet's video and image APIs at your fingertips.

Get started:
  gumlet login                          Authenticate with your API key
  gumlet video workspace list -o table  Browse your workspaces
  gumlet video asset upload --help      Upload a video

Use --output table for human-friendly tables, or --output json for scripts.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		ui.Error("%s", err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.SilenceUsage = true
	RootCmd.SilenceErrors = true
	RootCmd.PersistentFlags().String("output", "json", "Output format: json or table")
	RootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

func initConfig() {
	ui.Init(noColor)

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigName(".gumlet")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("GUMLET")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// Config loaded silently.
	}
}
