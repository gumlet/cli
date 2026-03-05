package video

import (
	"fmt"

	"github.com/spf13/cobra"
	"gumlet/pkg/client"
)

var getWorkspaceCmd = &cobra.Command{
	Use:   "get-workspace",
	Short: "Get details of a video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/workspaces/%s", workspaceID)
		resp, err := client.Get(path, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(getWorkspaceCmd)
	getWorkspaceCmd.Flags().String("workspace-id", "", "ID of the workspace")
	getWorkspaceCmd.MarkFlagRequired("workspace-id")
}
