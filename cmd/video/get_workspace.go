package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var getWorkspaceCmd = &cobra.Command{
	Use:   "get-workspace",
	Short: "Get details of a video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/workspaces/%s", workspaceID)
		resp, err := apiClient.Get(path, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(getWorkspaceCmd)
	getWorkspaceCmd.Flags().String("workspace-id", "", "ID of the workspace")
	getWorkspaceCmd.MarkFlagRequired("workspace-id")
}
