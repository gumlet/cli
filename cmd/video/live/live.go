package live

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	format, _ := cmd.Flags().GetString("output")
	if format == "json" || len(bytes.TrimSpace(raw)) == 0 {
		return printer.Print(cmd, raw, table)
	}
	if len(table.Columns) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Request succeeded. Use --output json to see the response; it may contain credentials.")
		return err
	}
	safe, err := tableResponse(raw, table)
	if err != nil {
		return err
	}
	return printer.Print(cmd, safe, table)
}

// Project live tables onto known scalar fields before sending them to the
// printer. Unexpected response shapes must not expose the unfiltered payload.
func tableResponse(raw []byte, table printer.Table) ([]byte, error) {
	var response map[string]json.RawMessage
	if json.Unmarshal(raw, &response) != nil || response == nil {
		return nil, errors.New("live API returned an unexpected response shape")
	}
	project := func(row map[string]json.RawMessage) map[string]json.RawMessage {
		safe := make(map[string]json.RawMessage)
		for _, column := range table.Columns {
			value := bytes.TrimSpace(row[column])
			if len(value) != 0 && !strings.ContainsRune("{[", rune(value[0])) {
				safe[column] = value
			}
		}
		return safe
	}
	if table.RowsKey == "" {
		return json.Marshal(project(response))
	}
	var rows []map[string]json.RawMessage
	if json.Unmarshal(response[table.RowsKey], &rows) != nil {
		return nil, errors.New("live API returned an unexpected list shape")
	}
	for i, row := range rows {
		rows[i] = project(row)
	}
	return json.Marshal(map[string]any{table.RowsKey: rows})
}
