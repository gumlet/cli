package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"gumlet/pkg/credentials"
	"gumlet/pkg/oauth"
	"gumlet/pkg/termui"

	"github.com/spf13/cobra"
)

func newLogoutCommand(store credentials.Store, oauthCfg oauth.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign out and remove saved credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stderr := cmd.ErrOrStderr()
			errStyle := termui.StyleFor(cmd, stderr)
			out := cmd.OutOrStdout()
			outStyle := termui.StyleFor(cmd, out)

			_, _ = fmt.Fprintf(stderr, "%s ", errStyle.Bold("Sign out and remove saved credentials?")+" "+errStyle.Dim("(y/N):"))
			scanner := bufio.NewScanner(cmd.InOrStdin())
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read confirmation: %w", err)
			}
			answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if answer != "y" && answer != "yes" {
				_, err := fmt.Fprintln(out, outStyle.Dim("Logout cancelled."))
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
					var spin *termui.Spinner
					if termui.IsTTY(stderr) {
						spin = termui.StartSpinner(stderr, "Revoking session...", errStyle.Color())
					}
					revokeErr := oauthCfg.Revoke(revokeCtx, tokenToRevoke, hint)
					spin.Stop()
					if revokeErr != nil {
						_, _ = fmt.Fprintln(stderr, errStyle.Warn(fmt.Sprintf("Warning: could not revoke token remotely: %v", revokeErr)))
					}
				}
			}

			removed, err := store.Remove()
			if err != nil {
				return err
			}
			if !removed {
				_, err = fmt.Fprintln(out, outStyle.Dim("No saved credentials."))
			} else {
				_, err = fmt.Fprintln(out, outStyle.Success("Logged out. Saved credentials removed."))
			}
			return err
		},
	}
}
