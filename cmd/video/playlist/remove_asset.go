package playlist

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var removeAssetCmd = &cobra.Command{
	Use:   "remove-asset",
	Short: "Remove assets from a playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		assetIDs, _ := cmd.Flags().GetStringSlice("asset-ids")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoPlaylists.DeleteAsset(cmd.Context(), playlistID, sdk.VideoPlaylistDeleteAssetParams{DeleteList: sdk.F(assetIDs)}, option.WithResponseBodyInto(&resp))
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
