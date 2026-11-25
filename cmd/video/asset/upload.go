package asset

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"
	"gumlet/pkg/termui"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/spf13/cobra"
)

func newUploadCommand(newClient client.Factory) *cobra.Command {
	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a local video file as a new asset",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			workspaceID, _ := cmd.Flags().GetString("workspace-id")
			format, _ := cmd.Flags().GetString("format")
			if format != "ABR" && format != "MP4" {
				return fmt.Errorf("invalid --format: use ABR or MP4")
			}
			file, info, err := openUploadFile(filePath, "video")
			if err != nil {
				return err
			}
			defer file.Close()

			params := sdk.VideoAssetUploadParams{
				WorkspaceID: sdk.String(workspaceID),
				Format:      sdk.F(sdk.VideoAssetUploadParamsFormat(format)),
				Title:       sdk.String(filepath.Base(filePath)),
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

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			stderr := cmd.ErrOrStderr()
			color := termui.ColorEnabled(cmd, stderr)

			sessionSpin := termui.StartSpinner(stderr, "Creating upload session...", color)
			result, err := apiClient.VideoAssets.Upload(cmd.Context(), params)
			if err != nil {
				sessionSpin.Fail("Could not create upload session")
				return err
			}

			if result == nil || result.AssetID == "" {
				sessionSpin.Fail("Upload session missing asset ID")
				return fmt.Errorf("upload session response did not include an asset ID")
			}
			assetID := result.AssetID
			if result.UploadURL == "" {
				sessionSpin.Fail("Upload session missing signed URL")
				return fmt.Errorf("asset %s was created, but the API did not return a signed upload URL", assetID)
			}
			sessionSpin.Succeed("Upload session ready")

			if err := uploadFile(cmd, result.UploadURL, filePath, file, info.Size(), "video", color); err != nil {
				return fmt.Errorf("asset %s was created, but upload failed: %w", assetID, err)
			}

			data, err := json.Marshal(struct {
				AssetID  string `json:"asset_id"`
				Uploaded bool   `json:"uploaded"`
			}{AssetID: assetID, Uploaded: true})
			if err != nil {
				return err
			}
			return printer.Print(cmd, data, printer.Table{
				Columns: []string{"asset_id"},
				Result:  "Asset uploaded.",
			})
		},
	}
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
	return uploadCmd
}

func uploadFile(cmd *cobra.Command, signedURL, filePath string, file *os.File, size int64, kind string, color bool) error {
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	stderr := cmd.ErrOrStderr()
	bar := termui.NewProgressBar(stderr, "Uploading "+kind, size, color)
	var progress func(int64)
	if bar != nil {
		progress = bar.Update
	} else {
		fmt.Fprintf(stderr, "Uploading %s...\n", kind)
	}
	err := client.Upload(cmd.Context(), signedURL, file, size, contentType, progress)
	bar.Finish("Uploaded "+kind, err)
	return err
}

func openUploadFile(path, kind string) (*os.File, os.FileInfo, error) {
	// Avoid opening special files such as FIFOs that can block before Stat.
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("inspect %s: %w", kind, err)
	}
	if !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("--file must be a nonempty regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", kind, err)
	}
	info, err = file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("inspect opened %s: %w", kind, err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		file.Close()
		return nil, nil, fmt.Errorf("--file must be a nonempty regular file")
	}
	return file, info, nil
}
