package asset

import (
	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "Manage video assets",
	}
	cmd.AddCommand(newDeleteCommand(newClient), newGetCommand(newClient), newListCommand(newClient), newUploadCommand(newClient))
	cmd.AddCommand(newUpdateCommand(newClient), newListRecoverableCommand(newClient), newRecoverCommand(newClient),
		newBulkDeleteCommand(newClient), newBulkTagCommand(newClient))
	return cmd
}
