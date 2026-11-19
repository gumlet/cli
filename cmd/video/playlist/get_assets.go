package playlist

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var getAssetsCmd = &cobra.Command{
	Use:   "get-assets",
	Short: "Get assets in a playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, _ := cmd.Flags().GetString("playlist-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoPlaylists.ListAssets(cmd.Context(), playlistID, sdk.VideoPlaylistListAssetsParams{}, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "title", "status", "created_at", "duration")
	},
}

func init() {
	Cmd.AddCommand(getAssetsCmd)
	getAssetsCmd.Flags().String("playlist-id", "", "ID of the playlist")
	getAssetsCmd.MarkFlagRequired("playlist-id")
}
