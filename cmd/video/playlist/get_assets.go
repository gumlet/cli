package playlist

import (
	"time"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newGetAssetsCommand(newClient client.Factory) *cobra.Command {
	getAssetsCmd := &cobra.Command{
		Use:   "get-assets",
		Short: "Get assets in a playlist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			playlistID, _ := cmd.Flags().GetString("playlist-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoPlaylists.ListAssets(cmd.Context(), playlistID, sdk.VideoPlaylistListAssetsParams{}, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{
				RowsKey: "asset_list",
				Columns: []string{"id", "title", "status", "created_at", "duration"},
				Fields: map[string]printer.Field{
					"id":         {Label: "Asset ID"},
					"created_at": {Label: "Created", Format: printer.Date(time.DateOnly, nil)},
				},
				More: "has_next_page",
			})
		},
	}
	getAssetsCmd.Flags().String("playlist-id", "", "ID of the playlist")
	getAssetsCmd.MarkFlagRequired("playlist-id")
	return getAssetsCmd
}
