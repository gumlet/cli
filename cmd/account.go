package cmd

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newUserCommand(newClient client.Factory) *cobra.Command {
	user := &cobra.Command{Use: "user", Short: "Inspect the current user"}
	user.AddCommand(newAccountReadCommand(newClient, "get", "Get the current user's details",
		printer.Table{
			Columns: []string{"name", "email", "id", "roles", "timezone"},
			Fields:  map[string]printer.Field{"roles": {Format: printer.JoinStrings(", ")}},
		},
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
		newAccountReadCommand(newClient, "history", "List invoice history",
			printer.Table{
				RowsKey: "stripe_invoices",
				Columns: []string{"created", "status", "amount_due", "id"},
				Summary: []string{"unpaid_invoices"},
				Fields: map[string]printer.Field{
					"created":    {Label: "Date (UTC)", Format: printer.Date(time.DateOnly, func(n int64) time.Time { return time.Unix(n, 0) })},
					"amount_due": {Format: invoiceAmount},
					"id":         {Label: "Invoice ID"},
				},
			},
			func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
				_, err := api.Billing.ListInvoices(ctx, opt)
				return err
			}),
		newAccountReadCommand(newClient, "upcoming", "Get the upcoming invoice and current usage",
			printer.Table{
				Columns: []string{"invoice_start_date", "invoice_last_date", "next_invoice_date", "current_month_usage", "upcoming_invoice.subtotal"},
				Fields: map[string]printer.Field{
					"invoice_start_date":        {Label: "Billing period start"},
					"invoice_last_date":         {Label: "Billing period end"},
					"current_month_usage":       {Label: "Current usage (incl. tax)"},
					"upcoming_invoice.subtotal": {Label: "Subtotal (excl. tax)"},
				},
			},
			func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
				_, err := api.Billing.FetchUpcomingInvoice(ctx, opt)
				return err
			}),
	)
	billing.AddCommand(invoice)
	return billing
}

func invoiceAmount(value any, row map[string]any) string {
	number, ok := value.(json.Number)
	if !ok {
		return ""
	}
	amount, err := number.Int64()
	if err != nil {
		return ""
	}
	code, _ := row["currency"].(string)
	code = strings.ToUpper(code)
	digits := 0
	switch code {
	case "BIF", "CLP", "DJF", "GNF", "JPY", "KMF", "KRW", "MGA", "PYG", "RWF", "VND", "VUV", "XAF", "XOF", "XPF":
		// Stripe zero-decimal currencies.
	case "BHD", "JOD", "KWD", "OMR", "TND":
		digits = 3
	case "AED", "AUD", "BRL", "CAD", "CHF", "CNY", "CZK", "DKK", "EUR", "GBP", "HKD", "HUF", "IDR",
		"ILS", "INR", "ISK", "MXN", "MYR", "NGN", "NOK", "NZD", "PHP", "PLN", "RON", "SAR", "SEK", "SGD",
		"THB", "TRY", "TWD", "UGX", "USD", "ZAR":
		digits = 2
	default:
		unit := "minor units"
		if len(code) == 3 && strings.Trim(code, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") == "" {
			unit = code + " minor units"
		}
		return strconv.FormatInt(amount, 10) + " " + unit
	}
	text := strconv.FormatInt(amount, 10)
	if digits > 0 {
		sign := ""
		if strings.HasPrefix(text, "-") {
			sign, text = "-", text[1:]
		}
		if missing := digits + 1 - len(text); missing > 0 {
			text = strings.Repeat("0", missing) + text
		}
		point := len(text) - digits
		text = sign + text[:point] + "." + text[point:]
	}
	return code + " " + text
}

func newAccountReadCommand(newClient client.Factory, use, short string,
	table printer.Table,
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
			return printer.Print(cmd, raw, table)
		},
	}
}
