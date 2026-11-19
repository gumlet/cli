package live

import (
	"context"
	"errors"
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "live", Short: "Manage live streaming workspaces and assets"}
	cmd.AddCommand(newWorkspaceCommand(newClient), newAssetCommand(newClient))
	return cmd
}

func leaf(use, short string, run func(*cobra.Command) error) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return run(cmd)
	}}
}

func requireString(cmd *cobra.Command, value *string, name, help string) {
	cmd.Flags().StringVar(value, name, "", help)
	cmd.MarkFlagRequired(name)
}

type request func(context.Context, *client.Client, option.RequestOption) error

func runRequest(cmd *cobra.Command, newClient client.Factory, table printer.Table, call request) error {
	api, err := newClient()
	if err != nil {
		return err
	}
	var raw []byte
	if err := call(cmd.Context(), api, option.WithResponseBodyInto(&raw)); err != nil {
		// SDK errors can include the raw response, potentially containing stream credentials.
		var apiErr *sdk.Error
		switch {
		case errors.As(err, &apiErr):
			return fmt.Errorf("live API request failed (HTTP %d)", apiErr.StatusCode)
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("live API request: %w", context.DeadlineExceeded)
		case errors.Is(err, context.Canceled):
			return fmt.Errorf("live API request: %w", context.Canceled)
		default:
			return errors.New("live API request failed; check your connection and retry if safe")
		}
	}
	return printer.Print(cmd, raw, table)
}
