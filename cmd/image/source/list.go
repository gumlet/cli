package source

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newListCommand(newClient client.Factory) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all image sources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.ImageSources.List(cmd.Context(), sdk.ImageSourceListParams{}, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{RowsKey: "all_sources", Columns: []string{"id", "namespace", "type", "cname", "created_at"}})
		},
	}
	return listCmd
}
