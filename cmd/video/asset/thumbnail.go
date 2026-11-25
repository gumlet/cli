package asset

import (
	"encoding/json"
	"fmt"
	"math"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newThumbnailUploadCommand(newClient client.Factory) *cobra.Command {
	var assetID, filePath string
	cmd := &cobra.Command{
		Use:   "thumbnail-upload",
		Short: "Upload a local image file as an asset's thumbnail",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(assetID) == "" {
				return fmt.Errorf("--asset-id must not be empty")
			}
			file, info, err := openVideo(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			api, err := newClient()
			if err != nil {
				return err
			}
			stderr := cmd.ErrOrStderr()
			fmt.Fprintln(stderr, "Requesting thumbnail upload URL...")
			result, err := api.VideoAssets.ThumbnailUpload(cmd.Context(), assetID)
			if err != nil {
				return err
			}
			if result == nil || result.UploadURL == "" {
				return fmt.Errorf("thumbnail upload response did not include a signed upload URL")
			}

			contentType := mime.TypeByExtension(filepath.Ext(filePath))
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			fmt.Fprintln(stderr, "Uploading thumbnail...")
			progress := uploadProgress(stderr, info.Size())
			err = client.Upload(cmd.Context(), result.UploadURL, file, info.Size(), contentType, progress)
			if progress != nil {
				fmt.Fprintln(stderr)
			}
			if err != nil {
				return fmt.Errorf("thumbnail upload failed: %w", err)
			}

			respAssetID := result.AssetID
			if respAssetID == "" {
				respAssetID = assetID
			}
			data, err := json.Marshal(struct {
				AssetID            string `json:"asset_id"`
				Uploaded           bool   `json:"uploaded"`
				ThumbnailUpdatedAt int64  `json:"thumbnail_updated_at,omitempty"`
			}{
				AssetID:            respAssetID,
				Uploaded:           true,
				ThumbnailUpdatedAt: result.ThumbnailUpdatedAt,
			})
			if err != nil {
				return err
			}
			return printer.Print(cmd, data, printer.Table{
				Result:  "Thumbnail uploaded.",
				Columns: []string{"asset_id", "thumbnail_updated_at"},
				Fields: map[string]printer.Field{
					"thumbnail_updated_at": {Label: "Updated", Format: printer.Date(printer.UTCDateTime, time.UnixMilli)},
				},
			})
		},
	}
	cmd.Flags().StringVar(&assetID, "asset-id", "", "ID of the video asset")
	cmd.MarkFlagRequired("asset-id")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to the local image file to upload")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newThumbnailSelectCommand(newClient client.Factory) *cobra.Command {
	var assetID string
	var frameAtSecond float64
	cmd := &cobra.Command{
		Use:   "thumbnail-select",
		Short: "Select a video frame as an asset's thumbnail",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(assetID) == "" {
				return fmt.Errorf("--asset-id must not be empty")
			}
			if math.IsNaN(frameAtSecond) || math.IsInf(frameAtSecond, 0) || frameAtSecond < 0 {
				return fmt.Errorf("--frame-at-second must be a non-negative number")
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.VideoAssets.ThumbnailSelect(cmd.Context(), assetID, sdk.VideoAssetThumbnailSelectParams{
				FrameAtSecond: sdk.F(frameAtSecond),
			}, option.WithResponseBodyInto(&raw))
			if err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{
				Result:  "Thumbnail updated.",
				Columns: []string{"asset_id", "thumbnail_updated_at"},
				Fields: map[string]printer.Field{
					"thumbnail_updated_at": {Label: "Updated", Format: printer.Date(printer.UTCDateTime, time.UnixMilli)},
				},
			})
		},
	}
	cmd.Flags().StringVar(&assetID, "asset-id", "", "ID of the video asset")
	cmd.MarkFlagRequired("asset-id")
	cmd.Flags().Float64Var(&frameAtSecond, "frame-at-second", 0, "Timestamp in seconds to extract the thumbnail from")
	cmd.MarkFlagRequired("frame-at-second")
	return cmd
}
