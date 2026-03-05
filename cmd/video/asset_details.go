package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var assetDetailsCmd = &cobra.Command{
	Use:   "asset-details",
	Short: "Get details of a video asset",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		assetID, _ := cmd.Flags().GetString("asset-id")

		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/assets/%s/%s", workspaceID, assetID)
		resp, err := apiClient.Get(path, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "asset_id", "status", "created_at", "updated_at", "tag", "folder", "source_id")
	},
}

func init() {
	Cmd.AddCommand(assetDetailsCmd)
	assetDetailsCmd.Flags().String("workspace-id", "", "Workspace ID of the asset")
	assetDetailsCmd.MarkFlagRequired("workspace-id")
	assetDetailsCmd.Flags().String("asset-id", "", "ID of the asset")
	assetDetailsCmd.MarkFlagRequired("asset-id")
}
