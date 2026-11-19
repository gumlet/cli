package asset

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newDeleteCommand(newClient client.Factory) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a video asset",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assetID, _ := cmd.Flags().GetString("asset-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			err = apiClient.VideoAssets.Delete(cmd.Context(), assetID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Result: "Asset deleted."})
		},
	}
	deleteCmd.Flags().String("asset-id", "", "ID of the asset to delete")
	deleteCmd.MarkFlagRequired("asset-id")
	return deleteCmd
}
