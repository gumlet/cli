package asset

import (
	"fmt"
	"time"

	"gumlet/cmd/video/internal/pagination"
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newListCommand(newClient client.Factory) *cobra.Command {
	var workspaceID, status, tag, title, folderID, playlistID, sortBy, orderBy string
	var page pagination.Page
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List assets in a video workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := page.Validate(cmd); err != nil {
				return err
			}
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
			if page.Number != 0 {
				params.PageNumber = sdk.Int(page.Number)
			}
			if page.Size != 0 {
				params.PageSize = sdk.Int(page.Size)
			}
			if page.Token != "" {
				if title != "" || (sortBy != "" && sortBy != "created_at") {
					return fmt.Errorf("--next-page-token requires the default or created_at sort and no title filter")
				}
				params.NextPageToken = sdk.String(page.Token)
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
	page.Bind(listCmd)
	return listCmd
}
