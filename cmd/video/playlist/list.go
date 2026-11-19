package playlist

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newListCommand(newClient client.Factory) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all playlists",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			workspaceID, _ := cmd.Flags().GetString("workspace-id")
			params := sdk.VideoPlaylistListAllParams{WorkspaceID: sdk.String(workspaceID)}

			var resp []byte
			_, err = apiClient.VideoPlaylists.ListAll(cmd.Context(), params, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{
				Columns: []string{"id", "title"},
				Fields:  map[string]printer.Field{"id": {Label: "Playlist ID"}},
			})
		},
	}
	listCmd.Flags().String("workspace-id", "", "Workspace ID to list playlists for")
	listCmd.MarkFlagRequired("workspace-id")
	return listCmd
}
