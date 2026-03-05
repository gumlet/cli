package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var updatePlaylistCmd = &cobra.Command{
	Use:   "update-playlist",
	Short: "Update a playlist",
	Run: func(cmd *cobra.Command, args []string) {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			body["title"] = v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			body["description"] = v
		}
		if cmd.Flags().Changed("channel-visibility") {
			v, _ := cmd.Flags().GetString("channel-visibility")
			body["channel_visibility"] = v
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/playlist/%s", playlistID)
		resp, err := apiClient.Post(path, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "id", "collection_id", "title", "description")
	},
}

func init() {
	Cmd.AddCommand(updatePlaylistCmd)
	updatePlaylistCmd.Flags().String("playlist-id", "", "ID of the playlist to update")
	updatePlaylistCmd.MarkFlagRequired("playlist-id")
	updatePlaylistCmd.Flags().String("title", "", "New title for the playlist")
	updatePlaylistCmd.Flags().String("description", "", "New description for the playlist")
	updatePlaylistCmd.Flags().String("channel-visibility", "", "Channel visibility setting")
}
