package source

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all image sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		resp, err := apiClient.Get("/image/sources", nil)
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "namespace", "type", "cname", "created_at")
	},
}

func init() {
	Cmd.AddCommand(listCmd)
}
