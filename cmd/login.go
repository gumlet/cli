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
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gumlet/pkg/ui"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Gumlet CLI",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		ui.PrintLoginBanner()
		fmt.Println()
		fmt.Fprintf(os.Stdout, "%s ", ui.Prompt("API Key:"))
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("reading API key: %w", err)
		}
		apiKey := string(bytePassword)
		fmt.Println()

		viper.Set("api-key", apiKey)

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("finding home directory: %w", err)
		}

		configFile := home + "/.gumlet.yaml"
		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("writing config file: %w", err)
		}

		fmt.Println()
		ui.Success("Authenticated! API key saved to %s", configFile)
		ui.Hint("Try: gumlet video workspace list --output table")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
