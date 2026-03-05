package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var deleteWorkspaceCmd = &cobra.Command{
	Use:   "delete-workspace",
	Short: "Delete a video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/workspaces/%s", workspaceID)
		resp, err := apiClient.Delete(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(deleteWorkspaceCmd)
	deleteWorkspaceCmd.Flags().String("workspace-id", "", "ID of the workspace to delete")
	deleteWorkspaceCmd.MarkFlagRequired("workspace-id")
}
