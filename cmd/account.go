package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
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

func newOrgCommand(newClient client.Factory) *cobra.Command {
	org := &cobra.Command{Use: "org", Short: "Inspect organization details"}
	org.AddCommand(newAccountReadCommand(newClient, "get", "Get organization details",
		printer.Table{
			Columns: []string{"id", "org_email", "plan", "plan_cycle", "unpaid", "sso_enabled", "created_at"},
			Summary: []string{"plan_data.video.display_name", "plan_data.image.display_name"},
			Fields: map[string]printer.Field{
				"org_email":                    {Label: "Email"},
				"plan_cycle":                   {Label: "Billing cycle"},
				"sso_enabled":                  {Label: "SSO enabled"},
				"created_at":                   {Label: "Created", Format: printer.Date(printer.UTCDateTime, time.UnixMilli)},
				"plan_data.video.display_name": {Label: "Video plan"},
				"plan_data.image.display_name": {Label: "Image plan"},
			},
		},
		func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.OrganizationData.FetchOrg(ctx, opt)
			return err
		}))
	return org
}

func newAuditLogCommand(newClient client.Factory) *cobra.Command {
	group := &cobra.Command{Use: "audit-log", Short: "Inspect organization audit logs"}
	var startAt, endAt string
	var pageNumber, pageSize int64
	var userEmails, activityTypes []string
	list := &cobra.Command{
		Use:   "list",
		Short: "Fetch organization audit logs (owner/admin only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			start, err := time.Parse(time.DateOnly, startAt)
			if err != nil {
				return fmt.Errorf("--start-at must be a date in YYYY-MM-DD format")
			}
			end, err := time.Parse(time.DateOnly, endAt)
			if err != nil {
				return fmt.Errorf("--end-at must be a date in YYYY-MM-DD format")
			}
			if end.Before(start) {
				return fmt.Errorf("--end-at must not be before --start-at")
			}
			params := sdk.AuditLogFetchParams{
				DateRange: sdk.F(sdk.AuditLogFetchParamsDateRange{
					StartAt: sdk.F(start),
					EndAt:   sdk.F(end),
				}),
			}
			if cmd.Flags().Changed("page-number") {
				if pageNumber < 1 {
					return fmt.Errorf("--page-number must be at least 1")
				}
				params.PageNumber = sdk.F(pageNumber)
			}
			if cmd.Flags().Changed("page-size") {
				if pageSize < 10 || pageSize > 500 {
					return fmt.Errorf("--page-size must be between 10 and 500")
				}
				params.PageSize = sdk.F(pageSize)
			}
			if cmd.Flags().Changed("user-email") {
				if len(userEmails) == 0 {
					return fmt.Errorf("--user-email must include at least one email address")
				}
				for i, email := range userEmails {
					addr, err := mail.ParseAddress(email)
					if err != nil || addr.Name != "" || addr.Address != email {
						return fmt.Errorf("--user-email entry %d must be a bare email address", i+1)
					}
				}
				params.UserEmail = sdk.F(userEmails)
			}
			if cmd.Flags().Changed("activity-type") {
				if len(activityTypes) == 0 {
					return fmt.Errorf("--activity-type must include at least one activity type")
				}
				types := make([]sdk.AuditLogFetchParamsActivityType, 0, len(activityTypes))
				for _, t := range activityTypes {
					val := sdk.AuditLogFetchParamsActivityType(t)
					if !val.IsKnown() {
						return fmt.Errorf("invalid --activity-type %q", t)
					}
					types = append(types, val)
				}
				params.ActivityType = sdk.F(types)
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			if _, err := api.AuditLogs.Fetch(cmd.Context(), params, option.WithResponseBodyInto(&raw)); err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{
				RowsKey: "data",
				Columns: []string{"timestamp", "activity", "action_user", "status_code", "source.type", "source.id"},
				Fields: map[string]printer.Field{
					"timestamp":        {Label: "Time", Format: printer.Date(printer.UTCDateTime, time.UnixMilli)},
					"action_user":      {Label: "User"},
					"status_code":      {Label: "Status"},
					"source.type":      {Label: "Resource"},
					"source.id":        {Label: "Resource ID"},
					"pagination.total": {Label: "Total events"},
				},
				Summary: []string{"pagination.total"},
				HasMore: func(root map[string]any) bool {
					pag, _ := root["pagination"].(map[string]any)
					page, ok1 := pag["page"].(json.Number)
					totalPages, ok2 := pag["total_pages"].(json.Number)
					if !ok1 || !ok2 {
						return false
					}
					p, err1 := page.Int64()
					tp, err2 := totalPages.Int64()
					return err1 == nil && err2 == nil && p < tp
				},
			})
		},
	}
	list.Flags().StringVar(&startAt, "start-at", "", "Start date in YYYY-MM-DD format")
	list.MarkFlagRequired("start-at")
	list.Flags().StringVar(&endAt, "end-at", "", "End date in YYYY-MM-DD format")
	list.MarkFlagRequired("end-at")
	list.Flags().Int64Var(&pageNumber, "page-number", 1, "Page number (starts at 1)")
	list.Flags().Int64Var(&pageSize, "page-size", 100, "Page size (10–500)")
	list.Flags().StringSliceVar(&userEmails, "user-email", nil, "Comma-separated user emails to filter by")
	list.Flags().StringSliceVar(&activityTypes, "activity-type", nil, "Comma-separated activity types to filter by")
	group.AddCommand(list)
	return group
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
