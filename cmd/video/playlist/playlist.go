package playlist

import (
	"gumlet/cmd/video"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "playlist",
	Short: "Manage video playlists",
}

func init() {
	video.Cmd.AddCommand(Cmd)
}
