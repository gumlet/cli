package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and remove saved API key",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Are you sure you want to logout? (y/N): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			fmt.Println("Logout cancelled.")
			return
		}

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error finding home directory:", err)
			return
		}

		configFile := home + "/.gumlet.yaml"
		if err := os.Remove(configFile); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Already logged out.")
			} else {
				fmt.Println("Error removing config file:", err)
			}
			return
		}

		fmt.Println("Logged out. API key removed from", configFile)
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
