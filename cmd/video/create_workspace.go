package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var createWorkspaceCmd = &cobra.Command{
	Use:   "create-workspace",
	Short: "Create a new video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		body := map[string]interface{}{
			"name": name,
		}

		output, _ := cmd.Root().PersistentFlags().GetString("output")

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

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(createWorkspaceCmd)
	createWorkspaceCmd.Flags().String("name", "", "Name of the new workspace")
	createWorkspaceCmd.MarkFlagRequired("name")
}
