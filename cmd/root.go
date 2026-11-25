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

	"gumlet/cmd/image"
	"gumlet/cmd/video"
	"gumlet/pkg/client"
	"gumlet/pkg/credentials"

	"github.com/spf13/cobra"
)

type Options struct {
	Version, Commit, Date string
	Credentials           credentials.Store
	NewClient             client.Factory
}

func NewRootCommand(opts Options) *cobra.Command {
	newClient := opts.NewClient
	if newClient == nil {
		newClient = func() (*client.Client, error) {
			key, err := opts.Credentials.APIKey()
			if err != nil {
				return nil, err
			}
			return client.New(key, opts.Version), nil
		}
	}
	var output string
	root := &cobra.Command{
		Use:   "gumlet",
		Short: "Gumlet CLI — manage video and image assets from your terminal",
		Long: `Gumlet CLI puts Gumlet's video and image APIs at your fingertips.

Get started:
  gumlet login                          Save your API key
  gumlet video workspace list -o table  Browse your workspaces
  gumlet video asset upload --help      Upload a video

Use --output table for human-friendly tables, or --output json for scripts.`,
		Version:       opts.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if output != "json" && output != "table" {
				return fmt.Errorf("invalid --output %q: use json or table", output)
			}
			return nil
		},
	}
	root.PersistentFlags().StringVarP(&output, "output", "o", "json", "Output format: json or table")
	root.PersistentFlags().Bool("no-color", false, "Disable colored output")
	root.SetVersionTemplate(versionText(opts.Version, opts.Commit, opts.Date))
	root.AddCommand(newLoginCommand(opts.Credentials), newLogoutCommand(opts.Credentials),
		newVersionCommand(opts.Version, opts.Commit, opts.Date), image.NewCommand(newClient), video.NewCommand(newClient))
	root.AddCommand(newUserCommand(newClient), newOrgCommand(newClient), newAuditLogCommand(newClient), newBillingCommand(newClient))
	return root
}
