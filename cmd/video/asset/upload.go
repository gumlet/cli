package asset

import (
	"fmt"
	"path/filepath"

	"gumlet/pkg/client"
	"gumlet/pkg/ui"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a local video file as a new asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		format, _ := cmd.Flags().GetString("format")

		params := sdk.VideoAssetUploadParams{
			CollectionID: sdk.String(workspaceID),
			Format:       sdk.F(sdk.VideoAssetUploadParamsFormat(format)),
			Title:        sdk.String(filepath.Base(filePath)),
		}

		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			params.Title = sdk.String(v)
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			params.Description = sdk.String(v)
		}
		if cmd.Flags().Changed("profile-id") {
			v, _ := cmd.Flags().GetString("profile-id")
			params.ProfileID = sdk.String(v)
		}
		if cmd.Flags().Changed("playlist-id") {
			v, _ := cmd.Flags().GetString("playlist-id")
			params.PlaylistID = sdk.String(v)
		}
		if cmd.Flags().Changed("tag") {
			v, _ := cmd.Flags().GetStringSlice("tag")
			params.Tag = sdk.F(v)
		}

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		ui.Info("Creating upload session...")
		result, err := apiClient.VideoAssets.Upload(cmd.Context(), params)
		if err != nil {
			return err
		}

		if result == nil || result.UploadURL == "" {
			return fmt.Errorf("upload_url not found in response")
		}

		assetID := result.AssetID
		ui.Success("Asset created: %s", assetID)

		if err := client.PutFile(cmd.Context(), result.UploadURL, filePath); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}

		ui.Success("Upload complete! Asset ID: %s", assetID)
		ui.Hint("Check status: gumlet video asset get --asset-id %s", assetID)
		return nil
	},
}

func init() {
	Cmd.AddCommand(uploadCmd)
	uploadCmd.Flags().String("file", "", "Path to the local video file to upload")
	uploadCmd.MarkFlagRequired("file")
	uploadCmd.Flags().String("workspace-id", "", "Workspace (collection) ID")
	uploadCmd.MarkFlagRequired("workspace-id")
	uploadCmd.Flags().String("format", "ABR", "Transcode format: ABR or MP4 (default: ABR)")
	uploadCmd.Flags().String("title", "", "Asset title")
	uploadCmd.Flags().String("description", "", "Asset description")
	uploadCmd.Flags().String("profile-id", "", "Video profile ID")
	uploadCmd.Flags().String("playlist-id", "", "Add asset to this playlist")
	uploadCmd.Flags().StringSlice("tag", []string{}, "Tags to apply (comma-separated)")
}
