package source

import (
	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Manage image sources",
	}
	cmd.AddCommand(newAddCommand(newClient), newDeleteCommand(newClient), newGetCommand(newClient), newListCommand(newClient), newUpdateCommand(newClient))
	return cmd
}
