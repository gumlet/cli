package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gumlet/pkg/ui"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and remove saved API key",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(os.Stdout, "%s ", ui.Prompt("Log out and remove saved API key? (y/N):"))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			ui.Info("Logout cancelled.")
			return nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("finding home directory: %w", err)
		}

		configFile := home + "/.gumlet.yaml"
		if err := os.Remove(configFile); err != nil {
			if os.IsNotExist(err) {
				ui.Info("Already logged out.")
				return nil
			}
			return fmt.Errorf("removing config file: %w", err)
		}

		ui.Success("Logged out. API key removed from %s", configFile)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
