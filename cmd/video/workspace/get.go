package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of a video workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoWorkspaces.Get(cmd.Context(), workspaceID, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "name", "type", "created_at", "updated_at")
	},
}

func init() {
	Cmd.AddCommand(getCmd)
	getCmd.Flags().String("workspace-id", "", "ID of the workspace")
	getCmd.MarkFlagRequired("workspace-id")
}
