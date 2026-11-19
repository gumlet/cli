package cmd

import (
	"context"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newUserCommand(newClient client.Factory) *cobra.Command {
	user := &cobra.Command{Use: "user", Short: "Inspect the current user"}
	user.AddCommand(newAccountReadCommand(newClient, "get", "Get the current user's details",
		func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.UserData.Fetch(ctx, opt)
			return err
		}))
	return user
}

func newBillingCommand(newClient client.Factory) *cobra.Command {
	billing := &cobra.Command{Use: "billing", Short: "Inspect billing information"}
	invoice := &cobra.Command{Use: "invoice", Short: "Inspect invoices"}
	invoice.AddCommand(
		newAccountReadCommand(newClient, "history", "List invoice history and unpaid invoices",
			func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
				_, err := api.Billing.ListInvoices(ctx, opt)
				return err
			}),
		newAccountReadCommand(newClient, "upcoming", "Get the upcoming invoice and current usage",
			func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
				_, err := api.Billing.FetchUpcomingInvoice(ctx, opt)
				return err
			}),
	)
	billing.AddCommand(invoice)
	return billing
}

func newAccountReadCommand(newClient client.Factory, use, short string,
	request func(context.Context, *client.Client, option.RequestOption) error,
) *cobra.Command {
	return &cobra.Command{
		Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			if err := request(cmd.Context(), api, option.WithResponseBodyInto(&raw)); err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{})
		},
	}
}
