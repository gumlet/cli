package asset

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of a video asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		assetID, _ := cmd.Flags().GetString("asset-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoAssets.GetDetails(cmd.Context(), assetID, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "asset_id", "input.title", "output.playback_url", "output.thumbnail_url", "status", "created_at", "updated_at", "tag", "collection_id")
	},
}

func init() {
	Cmd.AddCommand(getCmd)
	getCmd.Flags().String("asset-id", "", "ID of the asset")
	getCmd.MarkFlagRequired("asset-id")
}
