package workspace

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all video workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := apiClient.Get("/video/workspaces", nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "name", "type", "created_at")
	},
}

func init() {
	Cmd.AddCommand(listCmd)
}
