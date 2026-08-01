package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Gumlet CLI",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(os.Stdout, "Enter API Key: ")
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

		fmt.Println("Login successful. API key saved to", configFile)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
