package playlist

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
		Short: "Update a playlist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			playlistID, _ := cmd.Flags().GetString("playlist-id")

			params := sdk.VideoPlaylistUpdateParams{}
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				params.Title = sdk.String(v)
			}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				params.Description = sdk.String(v)
			}
			if cmd.Flags().Changed("channel-visibility") {
				v, _ := cmd.Flags().GetBool("channel-visibility")
				params.ChannelVisibility = sdk.Bool(v)
			}

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoPlaylists.Update(cmd.Context(), playlistID, params, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Columns: []string{"id", "workspace_id", "title", "description"}})
		},
	}
	updateCmd.Flags().String("playlist-id", "", "ID of the playlist to update")
	updateCmd.MarkFlagRequired("playlist-id")
	updateCmd.Flags().String("title", "", "New title for the playlist")
	updateCmd.Flags().String("description", "", "New description for the playlist")
	updateCmd.Flags().Bool("channel-visibility", false, "Show the playlist on the channel page (true or false)")
	updateCmd.MarkFlagsOneRequired("title", "description", "channel-visibility")
	return updateCmd
}
