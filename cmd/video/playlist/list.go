package playlist

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all playlists",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		collectionID, _ := cmd.Flags().GetString("workspace-id")

		params := sdk.VideoPlaylistListAllParams{}
		if collectionID != "" {
			params.CollectionID = sdk.String(collectionID)
		}

		var resp []byte
		_, err = apiClient.VideoPlaylists.ListAll(cmd.Context(), params, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "collection_id", "title", "description")
	},
}

func init() {
	Cmd.AddCommand(listCmd)
	listCmd.Flags().String("workspace-id", "", "Workspace ID to list playlists for")
	listCmd.MarkFlagRequired("workspace-id")
}
