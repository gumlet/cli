package asset

import (
	"gumlet/cmd/video"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "asset",
	Short: "Manage video assets",
}

func init() {
	video.Cmd.AddCommand(Cmd)
}
