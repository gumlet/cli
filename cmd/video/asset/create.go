package asset

import (
	"fmt"
	"strings"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newCreateCommand(newClient client.Factory) *cobra.Command {
	var workspaceID, input, format, title, description, profileID, playlistID, folderID, resolution, metadataFile string
	var tags []string
	var mp4Access, audioOnly, enableDRM bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a video asset from a URL",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(workspaceID) == "" {
				return fmt.Errorf("--workspace-id must not be empty")
			}
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("--input must not be empty")
			}
			flags := cmd.Flags()
			params := sdk.VideoAssetNewParams{
				WorkspaceID: sdk.F(workspaceID),
				Input:       sdk.F(input),
			}
			if flags.Changed("format") || !flags.Changed("profile-id") {
				f := sdk.VideoAssetNewParamsFormat(format)
				if !f.IsKnown() {
					return fmt.Errorf("invalid --format: use ABR or MP4")
				}
				params.Format = sdk.F(f)
			}
			if flags.Changed("title") {
				params.Title = sdk.F(title)
			}
			if flags.Changed("description") {
				params.Description = sdk.F(description)
			}
			if flags.Changed("profile-id") {
				params.ProfileID = sdk.F(profileID)
			}
			if flags.Changed("playlist-id") {
				params.PlaylistID = sdk.F(playlistID)
			}
			if flags.Changed("folder-id") {
				params.Folder = sdk.F(folderID)
			}
			if flags.Changed("tag") {
				params.Tag = sdk.F(tags)
			}
			if flags.Changed("resolution") {
				params.Resolution = sdk.F(resolution)
			}
			if flags.Changed("mp4-access") {
				params.Mp4Access = sdk.F(mp4Access)
			}
			if flags.Changed("audio-only") {
				params.AudioOnly = sdk.F(audioOnly)
			}
			if flags.Changed("enable-drm") {
				params.EnableDrm = sdk.F(enableDRM)
			}

			var raw []byte
			opts := []option.RequestOption{option.WithResponseBodyInto(&raw)}
			if flags.Changed("metadata-file") {
				data, err := readUpdateJSON(cmd, "metadata-file", metadataFile, '{')
				if err != nil {
					return err
				}
				opts = append(opts, option.WithJSONSet("metadata", data))
			}

			api, err := newClient()
			if err != nil {
				return err
			}
			if _, err := api.VideoAssets.New(cmd.Context(), params, opts...); err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{
				Result:  "Asset created.",
				Columns: []string{"asset_id", "input.title", "status", "workspace_id"},
			})
		},
	}

	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "Workspace (collection) ID")
	cmd.MarkFlagRequired("workspace-id")
	cmd.Flags().StringVar(&input, "input", "", "URL or web address of the video file to ingest")
	cmd.MarkFlagRequired("input")
	cmd.Flags().StringVar(&format, "format", "ABR", "Transcode format: ABR or MP4 (default: ABR unless --profile-id is set)")
	cmd.Flags().StringVar(&title, "title", "", "Asset title")
	cmd.Flags().StringVar(&description, "description", "", "Asset description")
	cmd.Flags().StringVar(&profileID, "profile-id", "", "Video profile ID")
	cmd.Flags().StringVar(&playlistID, "playlist-id", "", "Add asset to this playlist")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Add asset to an existing folder")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Comma-separated tags to apply")
	cmd.Flags().StringVar(&resolution, "resolution", "", "Comma-separated ABR resolutions (240p, 360p, 480p, 540p, 720p, 1080p)")
	cmd.Flags().BoolVar(&mp4Access, "mp4-access", false, "Create a downloadable MP4 version")
	cmd.Flags().BoolVar(&audioOnly, "audio-only", false, "Transcode and deliver audio-only")
	cmd.Flags().BoolVar(&enableDRM, "enable-drm", false, "Enable DRM encryption")
	cmd.Flags().StringVar(&metadataFile, "metadata-file", "", "Read metadata from a JSON object file (use - for stdin)")
	return cmd
}
