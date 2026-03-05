package video

import (
	"fmt"

	"github.com/spf13/cobra"
	"gumlet/pkg/client"
)

var deleteWorkspaceCmd = &cobra.Command{
	Use:   "delete-workspace",
	Short: "Delete a video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/workspaces/%s", workspaceID)
		resp, err := client.Delete(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(deleteWorkspaceCmd)
	deleteWorkspaceCmd.Flags().String("workspace-id", "", "ID of the workspace to delete")
	deleteWorkspaceCmd.MarkFlagRequired("workspace-id")
}
