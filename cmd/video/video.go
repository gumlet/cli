package video

import (
	"github.com/spf13/cobra"
	"gumlet/cmd/video/asset"
	"gumlet/cmd/video/playlist"
	"gumlet/cmd/video/workspace"
	"gumlet/pkg/client"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Manage video resources",
	}
	cmd.AddCommand(asset.NewCommand(newClient), playlist.NewCommand(newClient), workspace.NewCommand(newClient))
	return cmd
}
