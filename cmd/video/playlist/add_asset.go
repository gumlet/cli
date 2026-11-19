package playlist

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var addAssetCmd = &cobra.Command{
	Use:   "add-asset",
	Short: "Add assets to a playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		assetIDs, _ := cmd.Flags().GetStringSlice("asset-ids")
		positions, _ := cmd.Flags().GetIntSlice("positions")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		if len(assetIDs) == 0 {
			return fmt.Errorf("--asset-ids is required")
		}
		if len(positions) > 0 && len(positions) != len(assetIDs) {
			return fmt.Errorf("--positions count (%d) must match --asset-ids count (%d)", len(positions), len(assetIDs))
		}

		assetList := make([]sdk.VideoPlaylistNewAssetParamsAssetList, len(assetIDs))
		for i, id := range assetIDs {
			assetList[i].AssetID = sdk.String(id)
			if len(positions) > 0 {
				assetList[i].Position = sdk.Int(int64(positions[i]))
			}
		}

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoPlaylists.NewAsset(cmd.Context(), playlistID, sdk.VideoPlaylistNewAssetParams{AssetList: sdk.F(assetList)}, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output)
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
