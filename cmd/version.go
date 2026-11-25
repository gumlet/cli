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

	"gumlet/pkg/termui"

	"github.com/spf13/cobra"
)

func versionText(version, commit, date string) string {
	return fmt.Sprintf("gumlet version %s\ncommit: %s\nbuilt: %s\n", version, commit, date)
}

func newVersionCommand(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			style := termui.StyleFor(cmd, out)
			if !style.Color() {
				_, err := fmt.Fprint(out, versionText(version, commit, date))
				return err
			}
			_, err := fmt.Fprintf(out, "%s %s %s\n%s %s\n%s %s\n",
				style.BoldCyan("gumlet"),
				style.Dim("version"),
				style.Bold(version),
				style.Dim("commit:"),
				commit,
				style.Dim("built:"),
				date,
			)
			return err
		},
	}
}
