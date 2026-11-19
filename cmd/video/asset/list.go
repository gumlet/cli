package asset

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List assets in a video workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		opts := []option.RequestOption{option.WithResponseBodyInto(&resp)}
		// Preserve the existing list filters and their query keys during the SDK migration.
		for _, filter := range []string{"status", "tag", "title", "folder", "offset", "size", "playlist-id", "sort-by", "order-by", "type"} {
			if value, _ := cmd.Flags().GetString(filter); value != "" {
				opts = append(opts, option.WithQuery(filter, value))
			}
		}

		_, err = apiClient.VideoAssets.List(cmd.Context(), workspaceID, sdk.VideoAssetListParams{}, opts...)
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "asset_id", "input.title", "status", "created_at", "tag")
	},
}

func init() {
	Cmd.AddCommand(listCmd)
	listCmd.Flags().String("workspace-id", "", "Workspace ID to list assets from")
	listCmd.MarkFlagRequired("workspace-id")
	listCmd.Flags().String("status", "", "Filter by asset status")
	listCmd.Flags().String("tag", "", "Filter by asset tag")
	listCmd.Flags().String("title", "", "Filter by asset title")
	listCmd.Flags().String("folder", "", "Filter by asset folder")
	listCmd.Flags().String("offset", "", "Offset for pagination")
	listCmd.Flags().String("size", "", "Page size for pagination")
	listCmd.Flags().String("playlist-id", "", "Filter by playlist ID")
	listCmd.Flags().String("sort-by", "", "Field to sort by")
	listCmd.Flags().String("order-by", "", "Order to sort by")
	listCmd.Flags().String("type", "", "Type of asset to filter")
}
