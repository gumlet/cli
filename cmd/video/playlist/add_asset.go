package playlist

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var addAssetCmd = &cobra.Command{
	Use:   "add-asset",
	Short: "Add assets to a playlist",
	Run: func(cmd *cobra.Command, args []string) {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		assetIDs, _ := cmd.Flags().GetStringSlice("asset-ids")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"asset_list": assetIDs,
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/playlist/%s/asset", playlistID)
		resp, err := apiClient.Post(path, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(addAssetCmd)
	addAssetCmd.Flags().String("playlist-id", "", "ID of the playlist")
	addAssetCmd.MarkFlagRequired("playlist-id")
	addAssetCmd.Flags().StringSlice("asset-ids", []string{}, "Asset IDs to add to the playlist")
	addAssetCmd.MarkFlagRequired("asset-ids")
}
