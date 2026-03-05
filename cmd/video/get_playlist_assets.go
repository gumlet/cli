package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var getPlaylistAssetsCmd = &cobra.Command{
	Use:   "get-playlist-assets",
	Short: "Get assets in a playlist",
	Run: func(cmd *cobra.Command, args []string) {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/playlist/%s/assets", playlistID)
		resp, err := apiClient.Get(path, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "asset_id", "input.title", "status", "created_at")
	},
}

func init() {
	Cmd.AddCommand(getPlaylistAssetsCmd)
	getPlaylistAssetsCmd.Flags().String("playlist-id", "", "ID of the playlist")
	getPlaylistAssetsCmd.MarkFlagRequired("playlist-id")
}
