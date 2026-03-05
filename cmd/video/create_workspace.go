package video

import (
	"fmt"

	"github.com/spf13/cobra"
	"gumlet/pkg/client"
)

var createWorkspaceCmd = &cobra.Command{
	Use:   "create-workspace",
	Short: "Create a new video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		body := map[string]interface{}{
			"name": name,
		}

		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := client.Post("/video/workspaces", body)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(createWorkspaceCmd)
	createWorkspaceCmd.Flags().String("name", "", "Name of the new workspace")
	createWorkspaceCmd.MarkFlagRequired("name")
}
