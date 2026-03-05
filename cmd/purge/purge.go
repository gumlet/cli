package purge

import (
	cmd "gumlet/cmd"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge resources",
}

func init() {
	cmd.RootCmd.AddCommand(Cmd)
}
