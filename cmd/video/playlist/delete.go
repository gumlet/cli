package playlist

import (
	"fmt"
	"os"

	"gumlet/pkg/client"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, _ := cmd.Flags().GetString("playlist-id")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		err = apiClient.VideoPlaylists.Delete(cmd.Context(), playlistID, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(os.Stdout, string(resp))
		return err
	},
}

func init() {
	Cmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("playlist-id", "", "ID of the playlist to delete")
	deleteCmd.MarkFlagRequired("playlist-id")
}
