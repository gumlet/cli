package asset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newUpdateCommand(newClient client.Factory) *cobra.Command {
	var assetID, title, description, tag, input, metadataFile, ctaFile string
	var removeSubtitles []string
	var reprocess bool
	cmd := &cobra.Command{
		Use: "update", Short: "Update a video's metadata or processing settings", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(assetID) == "" {
				return fmt.Errorf("--asset-id must not be empty")
			}
			flags := cmd.Flags()
			params := sdk.VideoAssetUpdateParams{AssetID: sdk.F(assetID)}
			if flags.Changed("title") {
				params.Title = sdk.F(title)
			}
			if flags.Changed("description") {
				params.Description = sdk.F(description)
			}
			if flags.Changed("tag") {
				params.Tag = sdk.F(tag)
			}
			if flags.Changed("input") {
				params.Input = sdk.F(input)
			}
			if flags.Changed("remove-subtitles") {
				params.RemoveSubtitles = sdk.F(removeSubtitles)
			}
			if flags.Changed("reprocess") {
				params.Reprocess = sdk.F(reprocess)
			}
			if metadataFile == "-" && ctaFile == "-" {
				return fmt.Errorf("--metadata-file and --call-to-actions-file cannot both read stdin")
			}
			if flags.Changed("metadata-file") {
				data, err := readUpdateJSON(cmd, "metadata-file", metadataFile, '{')
				if err != nil {
					return err
				}
				params.Metadata = sdk.F(string(data))
			}
			var raw []byte
			opts := []option.RequestOption{option.WithResponseBodyInto(&raw)}
			if flags.Changed("call-to-actions-file") {
				data, err := readUpdateJSON(cmd, "call-to-actions-file", ctaFile, '[')
				if err != nil {
					return err
				}
				var entries []json.RawMessage
				if err := json.Unmarshal(data, &entries); err != nil {
					return fmt.Errorf("decode --call-to-actions-file: %w", err)
				}
				for _, entry := range entries {
					if len(entry) == 0 || entry[0] != '{' {
						return fmt.Errorf("--call-to-actions-file must contain an array of JSON objects")
					}
				}
				// Preserve JSON number precision and explicitly supplied empty values.
				opts = append(opts, option.WithJSONSet("call_to_actions", data))
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			_, err = api.VideoAssets.Update(cmd.Context(), params, opts...)
			if err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{})
		},
	}
	cmd.Flags().StringVar(&assetID, "asset-id", "", "ID of the asset")
	cmd.MarkFlagRequired("asset-id")
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVar(&tag, "tag", "", "New comma-separated tags")
	cmd.Flags().StringVar(&input, "input", "", "Replacement input path in Gumlet storage")
	cmd.Flags().BoolVar(&reprocess, "reprocess", false, "Reprocess the video")
	cmd.Flags().StringSliceVar(&removeSubtitles, "remove-subtitles", nil, "Subtitle language codes to remove")
	cmd.Flags().StringVar(&metadataFile, "metadata-file", "", "Read metadata from a JSON object file (use - for stdin)")
	cmd.Flags().StringVar(&ctaFile, "call-to-actions-file", "", "Read calls to action from a JSON array file (use - for stdin)")
	cmd.MarkFlagsOneRequired("title", "description", "tag", "input", "reprocess", "remove-subtitles", "metadata-file", "call-to-actions-file")
	return cmd
}

func readUpdateJSON(cmd *cobra.Command, flag, path string, first byte) (json.RawMessage, error) {
	input := cmd.InOrStdin()
	if path != "-" {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open --%s: %w", flag, err)
		}
		defer file.Close()
		input = file
	}
	const maxBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(input, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read --%s: %w", flag, err)
	}
	if len(data) > maxBytes {
		return nil, fmt.Errorf("--%s input exceeds 1 MiB", flag)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != first || !json.Valid(data) {
		kind := "object"
		if first == '[' {
			kind = "array"
		}
		return nil, fmt.Errorf("--%s must contain a valid JSON %s", flag, kind)
	}
	return data, nil
}
