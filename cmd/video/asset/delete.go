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
	RunE: func(cmd *cobra.Command, args []string) error {
		assetID, _ := cmd.Flags().GetString("asset-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/video/assets/%s", assetID)
		resp, err := apiClient.Delete(path)
		if err != nil {
			return err
		}

		return printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("asset-id", "", "ID of the asset to delete")
	deleteCmd.MarkFlagRequired("asset-id")
}
