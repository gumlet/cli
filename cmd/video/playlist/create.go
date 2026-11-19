package playlist

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
		Short: "Create a new playlist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID, _ := cmd.Flags().GetString("workspace-id")
			title, _ := cmd.Flags().GetString("title")
			description, _ := cmd.Flags().GetString("description")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			params := sdk.VideoPlaylistNewParams{WorkspaceID: sdk.String(workspaceID), Title: sdk.String(title)}
			if cmd.Flags().Changed("description") {
				params.Description = sdk.String(description)
			}
			_, err = apiClient.VideoPlaylists.New(cmd.Context(), params, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{
				Columns: []string{"id", "title", "workspace_id", "description"},
				Fields:  map[string]printer.Field{"id": {Label: "Playlist ID"}},
				Result:  "Playlist created.",
			})
		},
	}
	createCmd.Flags().String("workspace-id", "", "Workspace (collection) ID for the playlist")
	createCmd.MarkFlagRequired("workspace-id")
	createCmd.Flags().String("title", "", "Title of the playlist")
	createCmd.MarkFlagRequired("title")
	createCmd.Flags().String("description", "", "Description of the playlist")
	return createCmd
}
