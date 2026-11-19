package asset

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
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

		var resp []byte
		err = apiClient.VideoAssets.Delete(cmd.Context(), assetID, option.WithResponseBodyInto(&resp))
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
