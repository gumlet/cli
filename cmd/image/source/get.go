package source

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of an image source",
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceID, _ := cmd.Flags().GetString("source-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.ImageSources.Get(cmd.Context(), sourceID, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "namespace", "type", "subdomain", "cname", "is_active", "created_at", "updated_at")
	},
}

func init() {
	Cmd.AddCommand(getCmd)
	getCmd.Flags().String("source-id", "", "ID of the image source")
	getCmd.MarkFlagRequired("source-id")
}
