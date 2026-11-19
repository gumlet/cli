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

func newAddCommand(newClient client.Factory) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Create a new image source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, _ := cmd.Flags().GetString("namespace")
			sourceType, _ := cmd.Flags().GetString("type")
			if !validProvider(sourceType) {
				return fmt.Errorf("--type must name an image source provider")
			}
			config, err := sourceConfig(cmd)
			if err != nil {
				return err
			}
			body := map[string]any{
				"namespace": namespace,
				"type":      sourceType,
			}

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
			_, err = apiClient.ImageSources.New(cmd.Context(), sdk.ImageSourceNewParams{},
				// Keep the CLI's dynamic provider config, including legacy provider keys.
				option.WithRequestBody("application/json", payload), option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Columns: []string{"id", "namespace", "type", "created_at"}})
		},
	}
	addCmd.Flags().String("namespace", "", "Subdomain for the image source (e.g. mycompany.gumlet.com)")
	addCmd.MarkFlagRequired("namespace")
	addCmd.Flags().String("type", "", "Source type: amazon, proxy, gcs, dostorage, wasabi, cloudinary, azure, linode, backblaze, cloudflare")
	addCmd.MarkFlagRequired("type")
	configFlags(addCmd)
	return addCmd
}
