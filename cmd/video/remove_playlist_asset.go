package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var removePlaylistAssetCmd = &cobra.Command{
	Use:   "remove-playlist-asset",
	Short: "Remove assets from a playlist",
	Run: func(cmd *cobra.Command, args []string) {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		assetIDs, _ := cmd.Flags().GetStringSlice("asset-ids")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"delete_list": assetIDs,
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/playlist/%s/asset", playlistID)
		resp, err := apiClient.DeleteWithBody(path, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(removePlaylistAssetCmd)
	removePlaylistAssetCmd.Flags().String("playlist-id", "", "ID of the playlist")
	removePlaylistAssetCmd.MarkFlagRequired("playlist-id")
	removePlaylistAssetCmd.Flags().StringSlice("asset-ids", []string{}, "Asset IDs to remove from the playlist")
	removePlaylistAssetCmd.MarkFlagRequired("asset-ids")
}
