package image

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge image cache for a source",
	RunE: func(cmd *cobra.Command, args []string) error {
		subdomain, _ := cmd.Flags().GetString("subdomain")
		urls, _ := cmd.Flags().GetStringSlice("urls")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.ImageSources.PurgeCache(cmd.Context(), subdomain, sdk.ImageSourcePurgeCacheParams{Paths: sdk.F(urls)}, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(purgeCmd)
	purgeCmd.Flags().String("subdomain", "", "Subdomain to purge cache for")
	purgeCmd.MarkFlagRequired("subdomain")
	purgeCmd.Flags().StringSlice("urls", []string{}, "URLs to purge")
}
