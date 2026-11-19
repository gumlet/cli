package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newGetCommand(newClient client.Factory) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a video workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID, _ := cmd.Flags().GetString("workspace-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoWorkspaces.Get(cmd.Context(), workspaceID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Columns: []string{"id", "name", "type", "created_at", "updated_at"}})
		},
	}
	getCmd.Flags().String("workspace-id", "", "ID of the workspace")
	getCmd.MarkFlagRequired("workspace-id")
	return getCmd
}
