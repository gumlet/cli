package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newUpdateCommand(newClient client.Factory) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a video workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID, _ := cmd.Flags().GetString("workspace-id")
			name, _ := cmd.Flags().GetString("name")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoWorkspaces.Update(cmd.Context(), workspaceID, sdk.VideoWorkspaceUpdateParams{Name: sdk.String(name)}, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{
				Result:  "Workspace updated.",
				Columns: []string{"id", "name", "type"},
			})
		},
	}
	updateCmd.Flags().String("workspace-id", "", "ID of the workspace to update")
	updateCmd.MarkFlagRequired("workspace-id")
	updateCmd.Flags().String("name", "", "New name for the workspace")
	updateCmd.MarkFlagRequired("name")
	return updateCmd
}
