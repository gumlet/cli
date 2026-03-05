package video

import (
	"fmt"

	"github.com/spf13/cobra"
	"gumlet/pkg/client"
)

var listWorkspacesCmd = &cobra.Command{
	Use:   "list-workspaces",
	Short: "List all video workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := client.Get("/video/workspaces", nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(listWorkspacesCmd)
}
