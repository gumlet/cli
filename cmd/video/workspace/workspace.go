package workspace

import (
	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage video workspaces",
	}
	cmd.AddCommand(newCreateCommand(newClient), newDeleteCommand(newClient), newGetCommand(newClient), newListCommand(newClient), newUpdateCommand(newClient))
	return cmd
}
