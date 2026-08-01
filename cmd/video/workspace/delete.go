package workspace

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a video workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/video/workspaces/%s", workspaceID)
		resp, err := apiClient.Delete(path)
		if err != nil {
			return err
		}

		return printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("workspace-id", "", "ID of the workspace to delete")
	deleteCmd.MarkFlagRequired("workspace-id")
}
