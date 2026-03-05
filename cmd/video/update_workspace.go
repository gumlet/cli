package video

import (
	"fmt"

	"github.com/spf13/cobra"
	"gumlet/pkg/client"
)

var updateWorkspaceCmd = &cobra.Command{
	Use:   "update-workspace",
	Short: "Update a video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		name, _ := cmd.Flags().GetString("name")

		body := map[string]interface{}{
			"name": name,
		}

		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/workspaces/%s", workspaceID)
		resp, err := client.Put(path, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(updateWorkspaceCmd)
	updateWorkspaceCmd.Flags().String("workspace-id", "", "ID of the workspace to update")
	updateWorkspaceCmd.MarkFlagRequired("workspace-id")
	updateWorkspaceCmd.Flags().String("name", "", "New name for the workspace")
	updateWorkspaceCmd.MarkFlagRequired("name")
}
