package playlist

import (
	"fmt"
	"strings"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newReorderCommand(newClient client.Factory) *cobra.Command {
	var playlistID, assetID, sortBy, sortOrder string
	var pageNumber, pageSize, position int64
	cmd := &cobra.Command{
		Use:   "reorder",
		Short: "Move an asset or sort a playlist",
		Example: `  gumlet video playlist reorder --playlist-id PLAYLIST_ID --asset-id ASSET_ID --page-number 1 --page-size 20 --asset-position 0
  gumlet video playlist reorder --playlist-id PLAYLIST_ID --sort-by title --sort-order asc`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(playlistID) == "" {
				return fmt.Errorf("--playlist-id must not be empty")
			}
			params := sdk.VideoPlaylistReorderAssetParams{}
			if cmd.Flags().Changed("sort-by") {
				by := sdk.VideoPlaylistReorderAssetParamsBodyVariant1SortBy(sortBy)
				order := sdk.VideoPlaylistReorderAssetParamsBodyVariant1SortOrder(sortOrder)
				if !by.IsKnown() {
					return fmt.Errorf("invalid --sort-by: use title or created_at")
				}
				if !order.IsKnown() {
					return fmt.Errorf("invalid --sort-order: use asc or desc")
				}
				params.Body = sdk.VideoPlaylistReorderAssetParamsBodyVariant1{SortBy: sdk.F(by), SortOrder: sdk.F(order)}
			} else {
				if strings.TrimSpace(assetID) == "" {
					return fmt.Errorf("--asset-id must not be empty")
				}
				if pageNumber < 1 {
					return fmt.Errorf("--page-number must be at least 1")
				}
				if pageSize < 1 || pageSize > 200 {
					return fmt.Errorf("--page-size must be between 1 and 200")
				}
				if position < 0 || position >= pageSize {
					return fmt.Errorf("--asset-position must be between 0 and --page-size minus 1")
				}
				params.Body = sdk.VideoPlaylistReorderAssetParamsBodyVariant0{
					AssetID: sdk.F(assetID), PageNumber: sdk.F(pageNumber), PageSize: sdk.F(pageSize), AssetPosition: sdk.F(position),
				}
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.VideoPlaylists.ReorderAsset(cmd.Context(), playlistID, params, option.WithResponseBodyInto(&raw))
			if err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{})
		},
	}
	cmd.Flags().StringVar(&playlistID, "playlist-id", "", "ID of the playlist")
	cmd.MarkFlagRequired("playlist-id")
	cmd.Flags().StringVar(&assetID, "asset-id", "", "ID of the asset to move")
	cmd.Flags().Int64Var(&pageNumber, "page-number", 0, "Playlist page number (starts at 1)")
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Playlist page size (1–200)")
	cmd.Flags().Int64Var(&position, "asset-position", 0, "Zero-based position within the page")
	cmd.Flags().StringVar(&sortBy, "sort-by", "", "Sort field: title or created_at")
	cmd.Flags().StringVar(&sortOrder, "sort-order", "", "Sort direction: asc or desc")
	cmd.MarkFlagsOneRequired("asset-id", "sort-by")
	cmd.MarkFlagsMutuallyExclusive("asset-id", "sort-by")
	cmd.MarkFlagsRequiredTogether("asset-id", "page-number", "page-size", "asset-position")
	cmd.MarkFlagsRequiredTogether("sort-by", "sort-order")
	return cmd
}
