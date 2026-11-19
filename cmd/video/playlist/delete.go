package playlist

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newDeleteCommand(newClient client.Factory) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a playlist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			playlistID, _ := cmd.Flags().GetString("playlist-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			err = apiClient.VideoPlaylists.Delete(cmd.Context(), playlistID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{Result: "Playlist deleted."})
		},
	}
	deleteCmd.Flags().String("playlist-id", "", "ID of the playlist to delete")
	deleteCmd.MarkFlagRequired("playlist-id")
	return deleteCmd
}
