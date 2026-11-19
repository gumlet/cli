package image

import (
	"gumlet/cmd/image/source"
	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Manage image sources",
	}
	cmd.AddCommand(newPurgeCommand(newClient), source.NewCommand(newClient))
	return cmd
}
