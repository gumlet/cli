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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Enter API Key: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("\nError reading API key:", err)
			return
		}
		apiKey := string(bytePassword)
		fmt.Println()

		viper.Set("api-key", apiKey)

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error finding home directory:", err)
			return
		}

		configFile := home + "/.gumlet.yaml"
		if err := viper.WriteConfigAs(configFile); err != nil {
			fmt.Println("Error writing config file:", err)
			return
		}

		fmt.Println("Login successful. API key saved to", configFile)
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
