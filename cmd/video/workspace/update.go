package workspace

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a video workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		name, _ := cmd.Flags().GetString("name")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"name": name,
		}

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/video/workspaces/%s", workspaceID)
		resp, err := apiClient.Post(path, body)
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "name", "type", "updated_at")
	},
}

func init() {
	Cmd.AddCommand(updateCmd)
	updateCmd.Flags().String("workspace-id", "", "ID of the workspace to update")
	updateCmd.MarkFlagRequired("workspace-id")
	updateCmd.Flags().String("name", "", "New name for the workspace")
	updateCmd.MarkFlagRequired("name")
}
