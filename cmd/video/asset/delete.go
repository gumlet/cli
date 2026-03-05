package asset

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a video asset",
	Run: func(cmd *cobra.Command, args []string) {
		assetID, _ := cmd.Flags().GetString("asset-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/assets/%s", assetID)
		resp, err := apiClient.Delete(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("asset-id", "", "ID of the asset to delete")
	deleteCmd.MarkFlagRequired("asset-id")
}
