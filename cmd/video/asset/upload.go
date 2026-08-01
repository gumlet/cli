package asset

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a local video file as a new asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		format, _ := cmd.Flags().GetString("format")

		body := map[string]interface{}{
			"collection_id": workspaceID,
			"format":        format,
		}

		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			body["title"] = v
		} else {
			body["title"] = filepath.Base(filePath)
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			body["description"] = v
		}
		if cmd.Flags().Changed("profile-id") {
			v, _ := cmd.Flags().GetString("profile-id")
			body["profile_id"] = v
		}
		if cmd.Flags().Changed("playlist-id") {
			v, _ := cmd.Flags().GetString("playlist-id")
			body["playlist_id"] = v
		}
		if cmd.Flags().Changed("tag") {
			v, _ := cmd.Flags().GetStringSlice("tag")
			body["tag"] = v
		}

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		fmt.Println("Creating upload session...")
		resp, err := apiClient.Post("/video/assets/upload", body)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp, &result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		uploadURL, ok := result["upload_url"].(string)
		if !ok || uploadURL == "" {
			return fmt.Errorf("upload_url not found in response")
		}

		assetID, _ := result["asset_id"].(string)
		fmt.Printf("Asset created: %s\n", assetID)
		fmt.Println("Uploading file...")

		if err := apiClient.PutFile(uploadURL, filePath); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}

		fmt.Printf("Upload complete. Asset ID: %s\n", assetID)
		fmt.Println("Run `gumlet video asset get --asset-id " + assetID + "` to check processing status.")
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
