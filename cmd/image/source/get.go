package source

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of an image source",
	Run: func(cmd *cobra.Command, args []string) {
		sourceID, _ := cmd.Flags().GetString("source-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/image/sources/%s", sourceID)
		resp, err := apiClient.Get(path, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "namespace", "type", "subdomain", "cname", "is_active", "created_at", "updated_at")
	},
}

func init() {
	Cmd.AddCommand(getCmd)
	getCmd.Flags().String("source-id", "", "ID of the image source")
	getCmd.MarkFlagRequired("source-id")
}
