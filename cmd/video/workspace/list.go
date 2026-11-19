package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all video workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoWorkspaces.List(cmd.Context(), sdk.VideoWorkspaceListParams{}, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "name", "type", "created_at")
	},
}

func init() {
	Cmd.AddCommand(listCmd)
}
