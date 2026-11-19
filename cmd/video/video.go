package video

import (
	"gumlet/cmd/video/asset"
	"gumlet/cmd/video/channel"
	"gumlet/cmd/video/live"
	"gumlet/cmd/video/playlist"
	"gumlet/cmd/video/workspace"
	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Manage video resources",
	}
	cmd.AddCommand(asset.NewCommand(newClient), playlist.NewCommand(newClient), workspace.NewCommand(newClient))
	cmd.AddCommand(channel.NewCommand(newClient), live.NewCommand(newClient))
	return cmd
}
