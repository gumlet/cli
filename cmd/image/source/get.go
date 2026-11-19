package source

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newGetCommand(newClient client.Factory) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of an image source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceID, _ := cmd.Flags().GetString("source-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.ImageSources.Get(cmd.Context(), sourceID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			timestamp := printer.Date(printer.UTCDateTime, nil)
			return printer.Print(cmd, resp, printer.Table{
				Columns: []string{"id", "namespace", "type", "subdomain", "is_active", "created_at", "updated_at"},
				Summary: []string{"cname", "temp_cname", "secure_urls", "browser_cache_time", "cdn_cache_time"},
				Fields: map[string]printer.Field{
					"is_active":          {Label: "Active"},
					"created_at":         {Label: "Created", Format: timestamp},
					"updated_at":         {Label: "Updated", Format: timestamp},
					"cname":              {Label: "Custom domains", Format: printer.JoinStrings(", ")},
					"temp_cname":         {Label: "Temporary CNAMEs", Format: printer.JoinStrings(", ")},
					"secure_urls":        {Label: "Signed URLs"},
					"browser_cache_time": {Label: "Browser cache (s)"},
					"cdn_cache_time":     {Label: "CDN cache (s)"},
				},
			})
		},
	}
	getCmd.Flags().String("source-id", "", "ID of the image source")
	getCmd.MarkFlagRequired("source-id")
	return getCmd
}
