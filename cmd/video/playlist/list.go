package playlist

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all playlists",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		collectionID, _ := cmd.Flags().GetString("collection-id")

		resp, err := apiClient.Get("/video/playlist", map[string]string{"collection_id": collectionID})
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "collection_id", "title", "description")
	},
}

func init() {
	Cmd.AddCommand(listCmd)
	listCmd.Flags().String("collection-id", "", "Collection (workspace) ID to list playlists for")
	listCmd.MarkFlagRequired("collection-id")
}
