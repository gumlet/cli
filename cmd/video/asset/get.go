package asset

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of a video asset",
	Run: func(cmd *cobra.Command, args []string) {
		assetID, _ := cmd.Flags().GetString("asset-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/assets/%s", assetID)
		resp, err := apiClient.Get(path, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "asset_id", "input.title", "output.playback_url", "output.thumbnail_url", "status", "created_at", "updated_at", "tag", "collection_id")
	},
}

func init() {
	Cmd.AddCommand(getCmd)
	getCmd.Flags().String("asset-id", "", "ID of the asset")
	getCmd.MarkFlagRequired("asset-id")
}
