package asset

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newGetCommand(newClient client.Factory) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a video asset",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assetID, _ := cmd.Flags().GetString("asset-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoAssets.GetDetails(cmd.Context(), assetID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Columns: []string{"asset_id", "input.title", "output.playback_url", "output.thumbnail_url", "status", "created_at", "updated_at", "tag", "workspace_id"}})
		},
	}
	getCmd.Flags().String("asset-id", "", "ID of the asset")
	getCmd.MarkFlagRequired("asset-id")
	return getCmd
}
