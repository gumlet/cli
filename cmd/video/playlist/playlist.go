package playlist

import (
	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playlist",
		Short: "Manage video playlists",
	}
	cmd.AddCommand(newAddAssetCommand(newClient), newCreateCommand(newClient), newDeleteCommand(newClient), newGetAssetsCommand(newClient), newListCommand(newClient), newRemoveAssetCommand(newClient), newUpdateCommand(newClient))
	return cmd
}
