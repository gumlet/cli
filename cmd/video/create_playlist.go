package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var createPlaylistCmd = &cobra.Command{
	Use:   "create-playlist",
	Short: "Create a new playlist",
	Run: func(cmd *cobra.Command, args []string) {
		collectionID, _ := cmd.Flags().GetString("workspace-id")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"collection_id": collectionID,
			"title":         title,
			"description":   description,
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		resp, err := apiClient.Post("/video/playlist", body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "collection_id", "title", "description")
	},
}

func init() {
	Cmd.AddCommand(createPlaylistCmd)
	createPlaylistCmd.Flags().String("workspace-id", "", "Workspace (collection) ID for the playlist")
	createPlaylistCmd.MarkFlagRequired("workspace-id")
	createPlaylistCmd.Flags().String("title", "", "Title of the playlist")
	createPlaylistCmd.MarkFlagRequired("title")
	createPlaylistCmd.Flags().String("description", "", "Description of the playlist")
}
