package source

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newDeleteCommand(newClient client.Factory) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an image source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceID, _ := cmd.Flags().GetString("source-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.ImageSources.Delete(cmd.Context(), sourceID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Result: "Image source deleted."})
		},
	}
	deleteCmd.Flags().String("source-id", "", "ID of the image source to delete")
	deleteCmd.MarkFlagRequired("source-id")
	return deleteCmd
}
