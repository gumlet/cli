package playlist

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var removeAssetCmd = &cobra.Command{
	Use:   "remove-asset",
	Short: "Remove assets from a playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		assetIDs, _ := cmd.Flags().GetStringSlice("asset-ids")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"delete_list": assetIDs,
		}

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/video/playlist/%s/asset", playlistID)
		resp, err := apiClient.DeleteWithBody(path, body)
		if err != nil {
			return err
		}

		return printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(removeAssetCmd)
	removeAssetCmd.Flags().String("playlist-id", "", "ID of the playlist")
	removeAssetCmd.MarkFlagRequired("playlist-id")
	removeAssetCmd.Flags().StringSlice("asset-ids", []string{}, "Asset IDs to remove from the playlist")
	removeAssetCmd.MarkFlagRequired("asset-ids")
}
