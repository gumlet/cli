package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var deleteAssetCmd = &cobra.Command{
	Use:   "delete-asset",
	Short: "Delete a video asset",
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
		resp, err := apiClient.Delete(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(deleteAssetCmd)
	deleteAssetCmd.Flags().String("workspace-id", "", "Workspace ID of the asset")
	deleteAssetCmd.MarkFlagRequired("workspace-id")
	deleteAssetCmd.Flags().String("asset-id", "", "ID of the asset to delete")
	deleteAssetCmd.MarkFlagRequired("asset-id")
}
