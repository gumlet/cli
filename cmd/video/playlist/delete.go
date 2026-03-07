package playlist

import (
	"fmt"

	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a playlist",
	Run: func(cmd *cobra.Command, args []string) {
		playlistID, _ := cmd.Flags().GetString("playlist-id")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/playlist/%s", playlistID)
		resp, err := apiClient.Delete(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("playlist-id", "", "ID of the playlist to delete")
	deleteCmd.MarkFlagRequired("playlist-id")
}
