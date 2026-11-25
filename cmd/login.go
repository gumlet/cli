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
	"gumlet/pkg/termui"

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
			stderr := cmd.ErrOrStderr()
			errStyle := termui.StyleFor(cmd, stderr)
			out := cmd.OutOrStdout()
			outStyle := termui.StyleFor(cmd, out)

			if strings.TrimSpace(os.Getenv(credentials.EnvKey)) != "" {
				msg := fmt.Sprintf("Note: %s is set in your environment and will take precedence over saved credentials.", credentials.EnvKey)
				_, _ = fmt.Fprintln(stderr, errStyle.Warn(msg))
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), loginTimeout)
			defer cancel()

			var spinner *termui.Spinner
			tok, err := oauthCfg.Login(ctx, oauth.LoginOptions{
				NoBrowser: noBrowser,
				OnAuthURL: func(authURL string, opened bool, openErr error) {
					styledURL := errStyle.Link(authURL)
					if opened {
						_, _ = fmt.Fprintf(stderr, "Opening browser to sign in to Gumlet...\n%s\n  %s\n",
							errStyle.Dim("If the browser did not open, visit:"), styledURL)
					} else {
						_, _ = fmt.Fprintf(stderr, "Open this URL in your browser to sign in to Gumlet:\n  %s\n", styledURL)
					}
					spinner = termui.StartSpinner(stderr, "Waiting for authentication callback...", errStyle.Color())
				},
			})
			if err != nil {
				spinner.Fail("Authentication failed")
				return err
			}
			spinner.Succeed("Authenticated with Gumlet")

			path, err := store.SaveToken(credentialsToken(tok))
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(out, "%s Credentials saved to %s\n%s\n",
				outStyle.Success(outStyle.Bold("Logged in!")),
				outStyle.Dim(path),
				outStyle.Hint("Try: "+outStyle.Cyan("gumlet user get -o table")),
			)
			return err
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the sign-in URL instead of opening the browser automatically")
	return cmd
}
