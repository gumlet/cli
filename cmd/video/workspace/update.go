package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a video workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		name, _ := cmd.Flags().GetString("name")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoWorkspaces.Update(cmd.Context(), workspaceID, sdk.VideoWorkspaceUpdateParams{Name: sdk.String(name)}, option.WithResponseBodyInto(&resp))
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
