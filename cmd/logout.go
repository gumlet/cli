package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"gumlet/pkg/credentials"
	"gumlet/pkg/oauth"

	"github.com/spf13/cobra"
)

func newLogoutCommand(store credentials.Store, oauthCfg oauth.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign out and remove saved credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Sign out and remove saved credentials? (y/N): ")
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

			if tok, err := store.LoadToken(); err == nil {
				revokeCtx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
				defer cancel()
				tokenToRevoke, hint := tok.RefreshToken, "refresh_token"
				if tokenToRevoke == "" {
					tokenToRevoke, hint = tok.AccessToken, "access_token"
				}
				if tokenToRevoke != "" {
					if revokeErr := oauthCfg.Revoke(revokeCtx, tokenToRevoke, hint); revokeErr != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not revoke token remotely: %v\n", revokeErr)
					}
				}
			}

			removed, err := store.Remove()
			if err != nil {
				return err
			}
			if !removed {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "No saved credentials.")
			} else {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Saved credentials removed.")
			}
			return err
		},
	}
}
