package video

import (
	"fmt"

	"github.com/spf13/cobra"
	"gumlet/pkg/client"
)

var deleteAssetCmd = &cobra.Command{
	Use:   "delete-asset",
	Short: "Delete a video asset",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		assetID, _ := cmd.Flags().GetString("asset-id")

		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/assets/%s/%s", workspaceID, assetID)
		resp, err := client.Delete(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(deleteAssetCmd)
	deleteAssetCmd.Flags().String("workspace-id", "", "Workspace ID of the asset")
	deleteAssetCmd.MarkFlagRequired("workspace-id")
	deleteAssetCmd.Flags().String("asset-id", "", "ID of the asset to delete")
	deleteAssetCmd.MarkFlagRequired("asset-id")
}
