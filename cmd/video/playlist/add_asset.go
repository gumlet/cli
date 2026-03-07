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
		positions, _ := cmd.Flags().GetIntSlice("positions")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		if len(assetIDs) == 0 {
			fmt.Println("Error: --asset-ids is required")
			return
		}
		if len(positions) > 0 && len(positions) != len(assetIDs) {
			fmt.Printf("Error: --positions count (%d) must match --asset-ids count (%d)\n", len(positions), len(assetIDs))
			return
		}

		assetList := make([]interface{}, len(assetIDs))
		for i, id := range assetIDs {
			entry := map[string]interface{}{"asset_id": id}
			if len(positions) > 0 {
				entry["position"] = positions[i]
			}
			assetList[i] = entry
		}

		body := map[string]interface{}{
			"asset_list": assetList,
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
	addAssetCmd.Flags().StringSlice("asset-ids", []string{}, "Comma-separated asset IDs to add")
	addAssetCmd.MarkFlagRequired("asset-ids")
	addAssetCmd.Flags().IntSlice("positions", []int{}, "Comma-separated positions corresponding to each asset ID (optional)")
}
