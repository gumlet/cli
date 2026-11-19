package image

import (
	"fmt"
	"net/url"
	"strings"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newPurgeCommand(newClient client.Factory) *cobra.Command {
	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge image cache for a source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceID, _ := cmd.Flags().GetString("source-id")
			urls, _ := cmd.Flags().GetStringSlice("urls")
			if strings.TrimSpace(sourceID) == "" {
				return fmt.Errorf("--source-id must not be empty")
			}
			if len(urls) == 0 {
				return fmt.Errorf("--urls must contain at least one image path")
			}
			for i, path := range urls {
				u, err := url.Parse(path)
				if err != nil || strings.TrimSpace(path) == "" || strings.HasPrefix(path, "//") ||
					strings.ContainsAny(path, "?#") || u.IsAbs() {
					return fmt.Errorf("--urls entry %d must be an image path without a query or fragment", i+1)
				}
			}

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.ImageSources.Purge(cmd.Context(), sourceID, sdk.ImageSourcePurgeParams{Paths: sdk.F(urls)}, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Result: "Image cache purge requested."})
		},
	}
	purgeCmd.Flags().String("source-id", "", "ID of the image source")
	purgeCmd.MarkFlagRequired("source-id")
	purgeCmd.Flags().StringSlice("urls", nil, "Comma-separated image paths without query strings")
	purgeCmd.MarkFlagRequired("urls")
	return purgeCmd
}
