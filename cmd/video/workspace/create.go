package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newCreateCommand(newClient client.Factory) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new video workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoWorkspaces.New(cmd.Context(), sdk.VideoWorkspaceNewParams{Name: sdk.String(name)}, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{
				Result:  "Workspace created.",
				Columns: []string{"id", "name", "type"},
			})
		},
	}
	createCmd.Flags().String("name", "", "Name of the new workspace")
	createCmd.MarkFlagRequired("name")
	return createCmd
}
