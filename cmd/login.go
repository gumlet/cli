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
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gumlet/pkg/credentials"
	"gumlet/pkg/oauth"

	"github.com/spf13/cobra"
)

const loginTimeout = 5 * time.Minute

func newLoginCommand(store credentials.Store, oauthCfg oauth.Config) *cobra.Command {
	var noBrowser bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in to Gumlet via browser (OAuth 2.0)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(os.Getenv(credentials.EnvKey)) != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Note: %s is set in your environment and will take precedence over saved credentials.\n", credentials.EnvKey)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), loginTimeout)
			defer cancel()

			tok, err := oauthCfg.Login(ctx, oauth.LoginOptions{
				NoBrowser: noBrowser,
				OnAuthURL: func(authURL string, opened bool, openErr error) {
					if opened {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Opening browser to sign in to Gumlet...\nIf the browser did not open, visit:\n  %s\n", authURL)
					} else {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Open this URL in your browser to sign in to Gumlet:\n  %s\n", authURL)
					}
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Waiting for authentication callback...")
				},
			})
			if err != nil {
				return err
			}

			path, err := store.SaveToken(credentials.Token{
				AccessToken:  tok.AccessToken,
				RefreshToken: tok.RefreshToken,
				TokenType:    tok.TokenType,
				ExpiresAt:    tok.ExpiresAt,
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Logged in! Credentials saved to %s\nTry: gumlet user get -o table\n", path)
			return err
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the sign-in URL instead of opening the browser automatically")
	return cmd
}
