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

			return printer.Print(cmd, resp, printer.Table{Columns: []string{"id", "namespace", "type", "subdomain", "cname", "is_active", "created_at", "updated_at"}})
		},
	}
	getCmd.Flags().String("source-id", "", "ID of the image source")
	getCmd.MarkFlagRequired("source-id")
	return getCmd
}
