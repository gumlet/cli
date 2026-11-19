package playlist

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		collectionID, _ := cmd.Flags().GetString("workspace-id")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			return err
		}

		var resp []byte
		_, err = apiClient.VideoPlaylists.New(cmd.Context(), sdk.VideoPlaylistNewParams{CollectionID: sdk.String(collectionID), Title: sdk.String(title), Description: sdk.String(description)}, option.WithResponseBodyInto(&resp))
		if err != nil {
			return err
		}

		return printer.Print(resp, output, "id", "collection_id", "title", "description")
	},
}

func init() {
	Cmd.AddCommand(createCmd)
	createCmd.Flags().String("workspace-id", "", "Workspace (collection) ID for the playlist")
	createCmd.MarkFlagRequired("workspace-id")
	createCmd.Flags().String("title", "", "Title of the playlist")
	createCmd.MarkFlagRequired("title")
	createCmd.Flags().String("description", "", "Description of the playlist")
}
