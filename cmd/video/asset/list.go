package asset

import (
	"fmt"
	"time"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newListCommand(newClient client.Factory) *cobra.Command {
	var workspaceID, status, tag, title, folderID, playlistID, sortBy, orderBy, token string
	var pageNumber, pageSize int64
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List assets in a video workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := sdk.VideoAssetListParams{}
			if status != "" {
				params.Status = sdk.String(status)
			}
			if tag != "" {
				params.Tag = sdk.String(tag)
			}
			if title != "" {
				params.Title = sdk.String(title)
			}
			if cmd.Flags().Changed("folder-id") {
				params.FolderID = sdk.String(folderID)
			}
			if playlistID != "" {
				params.PlaylistID = sdk.String(playlistID)
			}
			if sortBy != "" {
				sort := sdk.VideoAssetListParamsSortBy(sortBy)
				if !sort.IsKnown() {
					return fmt.Errorf("invalid --sort-by: use created_at, uploaded_at, title, duration or status")
				}
				params.SortBy = sdk.F(sort)
			}
			if orderBy != "" {
				order := sdk.VideoAssetListParamsOrderBy(orderBy)
				if !order.IsKnown() {
					return fmt.Errorf("invalid --order-by: use asc or desc")
				}
				params.OrderBy = sdk.F(order)
			}
			if cmd.Flags().Changed("page-number") {
				if pageNumber < 1 {
					return fmt.Errorf("--page-number must be at least 1")
				}
				params.PageNumber = sdk.Int(pageNumber)
			}
			if cmd.Flags().Changed("page-size") {
				if pageSize < 1 || pageSize > 100 {
					return fmt.Errorf("--page-size must be between 1 and 100")
				}
				params.PageSize = sdk.Int(pageSize)
			}
			if token != "" {
				if title != "" || (sortBy != "" && sortBy != "created_at") {
					return fmt.Errorf("--next-page-token requires the default or created_at sort and no title filter")
				}
				params.NextPageToken = sdk.String(token)
			}
			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoAssets.List(cmd.Context(), workspaceID, params, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{
				RowsKey: "all_assets",
				Columns: []string{"asset_id", "input.title", "status", "created_at"},
				Fields: map[string]printer.Field{
					"created_at":        {Label: "Created", Format: printer.Date(time.DateOnly, time.UnixMilli)},
					"total_asset_count": {Label: "Total assets"},
				},
				Summary: []string{"total_asset_count"},
				More:    "next_page_token",
			})
		},
	}
	listCmd.Flags().StringVar(&workspaceID, "workspace-id", "", "Workspace ID to list assets from")
	listCmd.MarkFlagRequired("workspace-id")
	listCmd.Flags().StringVar(&status, "status", "", "Filter by asset status")
	listCmd.Flags().StringVar(&tag, "tag", "", "Filter by asset tag")
	listCmd.Flags().StringVar(&title, "title", "", "Filter by asset title")
	listCmd.Flags().StringVar(&folderID, "folder-id", "", "Filter by folder ID; pass an empty value for unfiled assets")
	listCmd.Flags().StringVar(&playlistID, "playlist-id", "", "Filter by playlist ID")
	listCmd.Flags().StringVar(&sortBy, "sort-by", "", "Sort by created_at, uploaded_at, title, duration or status")
	listCmd.Flags().StringVar(&orderBy, "order-by", "", "Sort direction: asc or desc")
	listCmd.Flags().Int64Var(&pageNumber, "page-number", 1, "Page number (starts at 1)")
	listCmd.Flags().Int64Var(&pageSize, "page-size", 20, "Page size (1–100)")
	listCmd.Flags().StringVar(&token, "next-page-token", "", "Cursor returned by the previous page")
	listCmd.MarkFlagsMutuallyExclusive("page-number", "next-page-token")
	return listCmd
}
