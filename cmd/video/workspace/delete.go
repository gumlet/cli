package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newDeleteCommand(newClient client.Factory) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a video workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID, _ := cmd.Flags().GetString("workspace-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoWorkspaces.Delete(cmd.Context(), workspaceID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Result: "Workspace deleted."})
		},
	}
	deleteCmd.Flags().String("workspace-id", "", "ID of the workspace to delete")
	deleteCmd.MarkFlagRequired("workspace-id")
	return deleteCmd
}
