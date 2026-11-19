package source

import (
	"encoding/json"
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newUpdateCommand(newClient client.Factory) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update an image source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceID, _ := cmd.Flags().GetString("source-id")
			sourceType, _ := cmd.Flags().GetString("type")
			if !validProvider(sourceType) {
				return fmt.Errorf("--type must name an image source provider")
			}
			config, err := sourceConfig(cmd)
			if err != nil {
				return err
			}
			body := map[string]any{"type": sourceType}
			if config != nil {
				body[sourceType] = config
			}
			payload, err := json.Marshal(body)
			if err != nil {
				return err
			}

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.ImageSources.Update(cmd.Context(), sourceID, sdk.ImageSourceUpdateParams{},
				// Keep the CLI's dynamic provider config, including legacy provider keys.
				option.WithRequestBody("application/json", payload), option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Columns: []string{"id", "namespace", "type", "updated_at"}})
		},
	}
	updateCmd.Flags().String("source-id", "", "ID of the image source to update")
	updateCmd.MarkFlagRequired("source-id")
	updateCmd.Flags().String("type", "", "Source type: amazon, proxy, gcs, dostorage, wasabi, cloudinary, azure, linode, backblaze, cloudflare")
	updateCmd.MarkFlagRequired("type")
	configFlags(updateCmd)
	return updateCmd
}
