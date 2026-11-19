package asset

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newListRecoverableCommand(newClient client.Factory) *cobra.Command {
	var workspaceID string
	var pageNumber, pageSize int64
	cmd := &cobra.Command{
		Use: "list-recoverable", Short: "List deleted assets in a workspace's recycle bin", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(workspaceID) == "" {
				return fmt.Errorf("--workspace-id must not be empty")
			}
			params := sdk.RecycleBinListParams{WorkspaceID: sdk.F(workspaceID)}
			if cmd.Flags().Changed("page-number") {
				if pageNumber < 1 {
					return fmt.Errorf("--page-number must be at least 1")
				}
				params.PageNumber = sdk.F(pageNumber)
			}
			if cmd.Flags().Changed("page-size") {
				if pageSize < 1 || pageSize > 100 {
					return fmt.Errorf("--page-size must be between 1 and 100")
				}
				params.PageSize = sdk.F(pageSize)
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.RecycleBin.List(cmd.Context(), params, option.WithResponseBodyInto(&raw))
			if err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{
				RowsKey: "all_assets",
				Columns: []string{"asset_id", "title", "deleted_at"},
				Fields: map[string]printer.Field{
					"deleted_at":        {Label: "Deleted", Format: printer.Date(printer.UTCDateTime, time.UnixMilli)},
					"total_asset_count": {Label: "Total recoverable assets"},
				},
				Summary: []string{"total_asset_count"},
				HasMore: func(root map[string]any) bool {
					total, _ := root["total_asset_count"].(json.Number)
					n, err := total.Int64()
					return err == nil && n > 0 && (n-1)/pageSize >= pageNumber
				},
			})
		},
	}
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "ID of the workspace")
	cmd.MarkFlagRequired("workspace-id")
	cmd.Flags().Int64Var(&pageNumber, "page-number", 1, "Page number (starts at 1)")
	cmd.Flags().Int64Var(&pageSize, "page-size", 20, "Page size (1–100)")
	return cmd
}

func newRecoverCommand(newClient client.Factory) *cobra.Command {
	var assetID string
	cmd := &cobra.Command{
		Use: "recover", Short: "Recover a deleted video asset", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(assetID) == "" {
				return fmt.Errorf("--asset-id must not be empty")
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			err = api.RecycleBin.Recover(cmd.Context(), sdk.RecycleBinRecoverParams{AssetID: sdk.F(assetID)}, option.WithResponseBodyInto(&raw))
			if err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{Result: "Asset recovered."})
		},
	}
	cmd.Flags().StringVar(&assetID, "asset-id", "", "ID of the deleted asset")
	cmd.MarkFlagRequired("asset-id")
	return cmd
}
