package source

import (
	"gumlet/cmd/image"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "source",
	Short: "Manage image sources",
}

func init() {
	image.Cmd.AddCommand(Cmd)
}
