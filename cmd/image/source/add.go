package source

import (
	"encoding/json"
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new image source",
	Run: func(cmd *cobra.Command, args []string) {
		namespace, _ := cmd.Flags().GetString("namespace")
		sourceType, _ := cmd.Flags().GetString("type")
		configStr, _ := cmd.Flags().GetString("config")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"namespace": namespace,
			"type":      sourceType,
		}

		if configStr != "" {
			var configObj map[string]interface{}
			if err := json.Unmarshal([]byte(configStr), &configObj); err != nil {
				fmt.Println("Invalid --config JSON:", err)
				return
			}
			body[sourceType] = configObj
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := apiClient.Post("/image/sources", body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "namespace", "type", "created_at")
	},
}

func init() {
	Cmd.AddCommand(addCmd)
	addCmd.Flags().String("namespace", "", "Subdomain for the image source (e.g. mycompany.gumlet.com)")
	addCmd.MarkFlagRequired("namespace")
	addCmd.Flags().String("type", "", "Source type: amazon, proxy, gcs, dostorage, wasabi, cloudinary, azure, linode, backblaze, cloudflare")
	addCmd.MarkFlagRequired("type")
	addCmd.Flags().String("config", "", "JSON config for the source type (e.g. '{\"bucket_name\":\"my-bucket\",\"bucket_region\":\"us-east-1\"}')")
}
