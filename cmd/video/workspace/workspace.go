package workspace

import (
	"gumlet/cmd/video"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage video workspaces",
}

func init() {
	video.Cmd.AddCommand(Cmd)
}
