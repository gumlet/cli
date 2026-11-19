package playlist

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

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
			v, _ := cmd.Flags().GetString("channel-visibility")
			// Preserve the CLI's documented string visibility values.
			params.ChannelVisibility = sdk.Raw[bool](v)
		}

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoPlaylists.Update(cmd.Context(), playlistID, params, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "collection_id", "title", "description")
	},
}

func init() {
	Cmd.AddCommand(updateCmd)
	updateCmd.Flags().String("playlist-id", "", "ID of the playlist to update")
	updateCmd.MarkFlagRequired("playlist-id")
	updateCmd.Flags().String("title", "", "New title for the playlist")
	updateCmd.Flags().String("description", "", "New description for the playlist")
	updateCmd.Flags().String("channel-visibility", "", "Channel visibility setting")
}
