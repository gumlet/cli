package live

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"gumlet/cmd/video/internal/pagination"
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var assetTable = printer.Table{Columns: []string{"live_asset_id", "status", "vod_asset_id", "created_at", "updated_at"}}

func newAssetCommand(newClient client.Factory) *cobra.Command {
	group := &cobra.Command{Use: "asset", Short: "Manage live streams"}
	group.AddCommand(newAssetCreate(newClient), newAssetUpdate(newClient), newAssetList(newClient), newThumbnailURLs(newClient))
	group.AddCommand(
		newAssetAction(newClient, "get", "Get live stream status", assetTable,
			func(ctx context.Context, api *client.Client, id string, opt option.RequestOption) error {
				_, err := api.LiveAssets.GetStatus(ctx, id, opt)
				return err
			}),
		newAssetAction(newClient, "delete", "Delete a live stream", printer.Table{},
			func(ctx context.Context, api *client.Client, id string, opt option.RequestOption) error {
				_, err := api.LiveAssets.Delete(ctx, id, opt)
				return err
			}),
		newAssetAction(newClient, "start", "Start a live stream", printer.Table{},
			func(ctx context.Context, api *client.Client, id string, opt option.RequestOption) error {
				return api.LiveAssets.Start(ctx, id, opt)
			}),
		newAssetAction(newClient, "complete", "Complete a live stream", printer.Table{},
			func(ctx context.Context, api *client.Client, id string, opt option.RequestOption) error {
				_, err := api.LiveAssets.Complete(ctx, id, opt)
				return err
			}),
		newAssetAction(newClient, "history", "Get live stream status history", assetTable,
			func(ctx context.Context, api *client.Client, id string, opt option.RequestOption) error {
				_, err := api.LiveAssets.StatusHistory(ctx, id, opt)
				return err
			}),
	)
	return group
}

func newAssetAction(newClient client.Factory, use, short string, table printer.Table,
	call func(context.Context, *client.Client, string, option.RequestOption) error,
) *cobra.Command {
	var id string
	cmd := leaf(use, short, func(cmd *cobra.Command) error {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("--live-asset-id must not be empty")
		}
		return runRequest(cmd, newClient, table, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			return call(ctx, api, id, opt)
		})
	})
	requireString(cmd, &id, "live-asset-id", "ID of the live asset")
	return cmd
}

func newAssetCreate(newClient client.Factory) *cobra.Command {
	var workspaceID, resolution, title, orientation, startAt string
	var mp4 bool
	cmd := leaf("create", "Create a live stream", func(cmd *cobra.Command) error {
		if strings.TrimSpace(workspaceID) == "" {
			return fmt.Errorf("--live-workspace-id must not be empty")
		}
		resolutions := []string{"240p", "360p", "480p", "540p", "720p", "1080p"}
		if !slices.Contains(resolutions, resolution) {
			return fmt.Errorf("--resolution must be one of %s", strings.Join(resolutions, ", "))
		}
		params := sdk.LiveStreamAssetNewParams{LiveSourceID: sdk.F(workspaceID), Resolution: sdk.F(resolution)}
		if cmd.Flags().Changed("title") {
			params.Title = sdk.F(title)
		}
		if cmd.Flags().Changed("mp4-access") {
			params.Mp4Access = sdk.F(mp4)
		}
		if cmd.Flags().Changed("orientation") {
			value := sdk.LiveStreamAssetNewParamsOrientation(orientation)
			if !value.IsKnown() {
				return fmt.Errorf("--orientation must be landscape or portrait")
			}
			params.Orientation = sdk.F(value)
		}
		if cmd.Flags().Changed("start-at") {
			value, err := scheduledTime(startAt)
			if err != nil {
				return err
			}
			params.StartAt = sdk.F(value)
		}
		return runRequest(cmd, newClient, assetTable, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveAssets.New(ctx, params, opt)
			return err
		})
	})
	requireString(cmd, &workspaceID, "live-workspace-id", "ID of the live workspace")
	requireString(cmd, &resolution, "resolution", "HLS resolution (240p, 360p, 480p, 540p, 720p, or 1080p)")
	cmd.Flags().StringVar(&title, "title", "", "Stream title")
	cmd.Flags().BoolVar(&mp4, "mp4-access", false, "Create a downloadable MP4")
	cmd.Flags().StringVar(&orientation, "orientation", "", "Video orientation: landscape or portrait")
	cmd.Flags().StringVar(&startAt, "start-at", "", "Scheduled start time (RFC 3339)")
	return cmd
}

func newAssetUpdate(newClient client.Factory) *cobra.Command {
	var id, title, startAt string
	cmd := leaf("update", "Update a live stream's title or schedule", func(cmd *cobra.Command) error {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("--live-asset-id must not be empty")
		}
		params := sdk.LiveStreamAssetUpdateParams{LiveAssetID: sdk.F(id)}
		if cmd.Flags().Changed("title") {
			params.Title = sdk.F(title)
		}
		if cmd.Flags().Changed("start-at") {
			value, err := scheduledTime(startAt)
			if err != nil {
				return err
			}
			params.StartAt = sdk.F(value)
		}
		return runRequest(cmd, newClient, assetTable, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveAssets.Update(ctx, params, opt)
			return err
		})
	})
	requireString(cmd, &id, "live-asset-id", "ID of the live asset")
	cmd.Flags().StringVar(&title, "title", "", "New stream title")
	cmd.Flags().StringVar(&startAt, "start-at", "", "New scheduled start time (RFC 3339)")
	cmd.MarkFlagsOneRequired("title", "start-at")
	return cmd
}

func scheduledTime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("--start-at must be an RFC 3339 timestamp (for example 2026-12-01T15:00:00Z)")
	}
	return t, nil
}

func newAssetList(newClient client.Factory) *cobra.Command {
	var workspaceID string
	var statuses []string
	var page pagination.Page
	cmd := leaf("list", "List live streams in a workspace", func(cmd *cobra.Command) error {
		if strings.TrimSpace(workspaceID) == "" {
			return fmt.Errorf("--live-workspace-id must not be empty")
		}
		if err := page.Validate(cmd); err != nil {
			return err
		}
		params := sdk.LiveStreamAssetFilterParams{}
		if cmd.Flags().Changed("status") {
			if err := validateChoices(statuses, "status", []string{"created", "active", "complete", "disconnected", "errored", "deleted"}); err != nil {
				return err
			}
			params.Status = sdk.F(strings.Join(statuses, ","))
		}
		if page.Number != 0 {
			params.PageNumber = sdk.F(page.Number)
		}
		if page.Size != 0 {
			params.PageSize = sdk.F(page.Size)
		}
		if page.Token != "" {
			params.NextPageToken = sdk.F(page.Token)
		}
		table := assetTable
		table.RowsKey = "all_live_assets"
		return runRequest(cmd, newClient, table, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveAssets.Filter(ctx, workspaceID, params, opt)
			return err
		})
	})
	requireString(cmd, &workspaceID, "live-workspace-id", "ID of the live workspace")
	cmd.Flags().StringSliceVar(&statuses, "status", nil, "Comma-separated stream statuses")
	page.Bind(cmd)
	return cmd
}

func newThumbnailURLs(newClient client.Factory) *cobra.Command {
	var id string
	var statuses []string
	cmd := leaf("thumbnail-upload-urls", "Get presigned thumbnail upload URLs (JSON output only)", func(cmd *cobra.Command) error {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("--live-asset-id must not be empty")
		}
		if err := validateChoices(statuses, "statuses", []string{"preparing", "disconnected", "end"}); err != nil {
			return err
		}
		return runRequest(cmd, newClient, printer.Table{}, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveAssets.Upload(ctx, sdk.LiveStreamAssetUploadParams{
				LiveAssetID: sdk.F(id), Statuses: sdk.F[any](statuses),
			}, opt)
			return err
		})
	})
	requireString(cmd, &id, "live-asset-id", "ID of the live asset")
	cmd.Flags().StringSliceVar(&statuses, "statuses", nil, "Comma-separated thumbnail states: preparing, disconnected, end")
	cmd.MarkFlagRequired("statuses")
	return cmd
}

func validateChoices(values []string, flag string, allowed []string) error {
	if len(values) == 0 {
		return fmt.Errorf("--%s requires at least one value", flag)
	}
	for _, value := range values {
		if !slices.Contains(allowed, value) {
			return fmt.Errorf("--%s accepts only %s", flag, strings.Join(allowed, ", "))
		}
	}
	return nil
}
