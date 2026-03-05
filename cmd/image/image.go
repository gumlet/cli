package image

import (
	cmd "gumlet/cmd"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "image",
	Short: "Manage image sources",
}

func init() {
	cmd.RootCmd.AddCommand(Cmd)
}
