package video

import (
	"github.com/spf13/cobra"
	cmd "gumlet/cmd"
)

var Cmd = &cobra.Command{
	Use:   "video",
	Short: "Manage video resources",
}

func init() {
	cmd.RootCmd.AddCommand(Cmd)
}
