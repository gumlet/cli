package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"gumlet/pkg/credentials"

	"github.com/spf13/cobra"
)

func newLogoutCommand(store credentials.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved API key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Remove your saved API key? (y/N): ")
			scanner := bufio.NewScanner(cmd.InOrStdin())
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read confirmation: %w", err)
			}
			answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if answer != "y" && answer != "yes" {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "Logout cancelled.")
				return err
			}
			removed, err := store.Remove()
			if err != nil {
				return err
			}
			if !removed {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "No saved API key.")
			} else {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "Saved API key removed.")
			}
			return err
		},
	}
}
