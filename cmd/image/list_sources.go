package image

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var listSourcesCmd = &cobra.Command{
	Use:   "list-sources",
	Short: "List all image sources",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := apiClient.Get("/image/sources", nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "namespace", "type", "created_at")
	},
}

func init() {
	Cmd.AddCommand(listSourcesCmd)
}
