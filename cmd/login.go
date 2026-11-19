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

	"gumlet/pkg/credentials"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const apiKeysURL = "https://dash.gumlet.com/developer/api-keys"

func newLoginCommand(store credentials.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Save a Gumlet API key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Create an API key in the Gumlet dashboard:\n  %s\n", apiKeysURL)
			in, ok := cmd.InOrStdin().(*os.File)
			if !ok || !term.IsTerminal(int(in.Fd())) {
				return fmt.Errorf("login requires a terminal; for automation, set %s instead", credentials.EnvKey)
			}
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), "API key (hidden): ")
			key, err := term.ReadPassword(int(in.Fd()))
			_, _ = fmt.Fprintln(cmd.ErrOrStderr())
			if err != nil {
				return fmt.Errorf("read API key: %w", err)
			}
			path, err := store.Save(string(key))
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "API key saved to %s\nTry: gumlet video workspace list -o table\n", path)
			return err
		},
	}
}
