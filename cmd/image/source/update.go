package source

import (
	"encoding/json"
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an image source",
	Run: func(cmd *cobra.Command, args []string) {
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
				fmt.Println("Invalid --config JSON:", err)
				return
			}
			if sourceType != "" {
				body[sourceType] = configObj
			}
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/image/sources/%s", sourceID)
		resp, err := apiClient.Post(path, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "namespace", "type", "updated_at")
	},
}

func init() {
	Cmd.AddCommand(updateCmd)
	updateCmd.Flags().String("source-id", "", "ID of the image source to update")
	updateCmd.MarkFlagRequired("source-id")
	updateCmd.Flags().String("type", "", "Source type: amazon, proxy, gcs, dostorage, wasabi, cloudinary, azure, linode, backblaze, cloudflare")
	updateCmd.Flags().String("config", "", "JSON config for the source type")
}
