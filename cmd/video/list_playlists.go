package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var listPlaylistsCmd = &cobra.Command{
	Use:   "list-playlists",
	Short: "List all playlists",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := apiClient.Get("/video/playlist", nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "collection_id", "title", "description")
	},
}

func init() {
	Cmd.AddCommand(listPlaylistsCmd)
}
