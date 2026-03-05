package workspace

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"name": name,
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := apiClient.Post("/video/workspaces", body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "name", "type", "created_at")
	},
}

func init() {
	Cmd.AddCommand(createCmd)
	createCmd.Flags().String("name", "", "Name of the new workspace")
	createCmd.MarkFlagRequired("name")
}
