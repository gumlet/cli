package video

import (
	"fmt"

	"github.com/spf13/cobra"
	"gumlet/pkg/client"
)

var assetDetailsCmd = &cobra.Command{
	Use:   "asset-details",
	Short: "Get details of a video asset",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		assetID, _ := cmd.Flags().GetString("asset-id")

		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/assets/%s/%s", workspaceID, assetID)
		resp, err := client.Get(path, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(assetDetailsCmd)
	assetDetailsCmd.Flags().String("workspace-id", "", "Workspace ID of the asset")
	assetDetailsCmd.MarkFlagRequired("workspace-id")
	assetDetailsCmd.Flags().String("asset-id", "", "ID of the asset")
	assetDetailsCmd.MarkFlagRequired("asset-id")
}
