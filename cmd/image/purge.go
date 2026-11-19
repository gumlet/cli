package image

import (
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
			subdomain, _ := cmd.Flags().GetString("subdomain")
			urls, _ := cmd.Flags().GetStringSlice("urls")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.ImageSources.PurgeCache(cmd.Context(), subdomain, sdk.ImageSourcePurgeCacheParams{Paths: sdk.F(urls)}, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{})
		},
	}
	purgeCmd.Flags().String("subdomain", "", "Subdomain to purge cache for")
	purgeCmd.MarkFlagRequired("subdomain")
	purgeCmd.Flags().StringSlice("urls", []string{}, "Comma-separated paths or URLs without query parameters (omit to purge all)")
	return purgeCmd
}
