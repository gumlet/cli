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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an image source",
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceID, _ := cmd.Flags().GetString("source-id")
		sourceType, _ := cmd.Flags().GetString("type")
		configStr, _ := cmd.Flags().GetString("config")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{}
		if sourceType != "" {
			body["type"] = sourceType
		}
		if configStr != "" {
			var configObj map[string]interface{}
			if err := json.Unmarshal([]byte(configStr), &configObj); err != nil {
				return fmt.Errorf("invalid --config JSON: %w", err)
			}
			if sourceType != "" {
				body[sourceType] = configObj
			}
		}
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}

		apiClient, err := client.NewClient()
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

		return printer.Print(resp, output, "id", "namespace", "type", "updated_at")
	},
}

func init() {
	Cmd.AddCommand(updateCmd)
	updateCmd.Flags().String("source-id", "", "ID of the image source to update")
	updateCmd.MarkFlagRequired("source-id")
	updateCmd.Flags().String("type", "", "Source type: amazon, proxy, gcs, dostorage, wasabi, cloudinary, azure, linode, backblaze, cloudflare")
	updateCmd.Flags().String("config", "", "JSON config for the source type")
}
