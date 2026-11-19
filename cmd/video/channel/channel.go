package channel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"
	"unicode/utf8"

	"gumlet/cmd/video/internal/pagination"
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func NewCommand(newClient client.Factory) *cobra.Command {
	channel := &cobra.Command{Use: "channel", Short: "Manage video channels"}
	viewer := &cobra.Command{Use: "viewer", Short: "Manage members-only channel viewers"}
	viewer.AddCommand(newListCommand(newClient), newInviteCommand(newClient), newRemoveCommand(newClient))
	channel.AddCommand(viewer)
	return channel
}

func viewerCommand(use, short string, run func(*cobra.Command, string) error) *cobra.Command {
	var workspaceID string
	cmd := &cobra.Command{
		Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(workspaceID) == "" {
				return fmt.Errorf("--workspace-id must not be empty")
			}
			return run(cmd, workspaceID)
		},
	}
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "ID of the video workspace")
	cmd.MarkFlagRequired("workspace-id")
	return cmd
}

func newListCommand(newClient client.Factory) *cobra.Command {
	var page pagination.Page
	cmd := viewerCommand("list", "List channel subscribers", func(cmd *cobra.Command, workspaceID string) error {
		if err := page.Validate(cmd); err != nil {
			return err
		}
		params := sdk.ChannelViewerListSubscribersParams{}
		if page.Number != 0 {
			params.PageNumber = sdk.F(page.Number)
		}
		if page.Size != 0 {
			params.PageSize = sdk.F(page.Size)
		}
		if page.Token != "" {
			params.NextPageToken = sdk.F(page.Token)
		}
		api, err := newClient()
		if err != nil {
			return err
		}
		var raw []byte
		_, err = api.ChannelViewers.ListSubscribers(cmd.Context(), workspaceID, params, option.WithResponseBodyInto(&raw))
		if err != nil {
			return err
		}
		return printer.Print(cmd, raw, printer.Table{RowsKey: "subscriptions"})
	})
	page.Bind(cmd)
	return cmd
}

func newInviteCommand(newClient client.Factory) *cobra.Command {
	var file string
	cmd := viewerCommand("invite", "Invite viewers to a members-only channel", func(cmd *cobra.Command, workspaceID string) error {
		users, err := readViewers(cmd, file)
		if err != nil {
			return err
		}
		api, err := newClient()
		if err != nil {
			return err
		}
		var raw []byte
		_, err = api.ChannelViewers.Invite(cmd.Context(), workspaceID, sdk.ChannelViewerInviteParams{
			Users: sdk.F(users),
		}, option.WithResponseBodyInto(&raw))
		if err != nil {
			return err
		}
		return printer.Print(cmd, raw, printer.Table{})
	})
	cmd.Flags().StringVar(&file, "viewers-file", "", "JSON file of {email,name} entries (use - for stdin)")
	cmd.MarkFlagRequired("viewers-file")
	return cmd
}

func newRemoveCommand(newClient client.Factory) *cobra.Command {
	var emails []string
	cmd := viewerCommand("remove", "Remove subscribers from a channel", func(cmd *cobra.Command, workspaceID string) error {
		if len(emails) < 1 || len(emails) > 200 {
			return fmt.Errorf("--emails must contain 1–200 email addresses")
		}
		for i, email := range emails {
			if !validEmail(email) {
				return fmt.Errorf("--emails entry %d must be a bare email address", i+1)
			}
		}
		api, err := newClient()
		if err != nil {
			return err
		}
		var raw []byte
		_, err = api.ChannelViewers.Remove(cmd.Context(), workspaceID, sdk.ChannelViewerRemoveParams{
			Emails: sdk.F(emails),
		}, option.WithResponseBodyInto(&raw))
		if err != nil {
			return err
		}
		return printer.Print(cmd, raw, printer.Table{})
	})
	cmd.Flags().StringSliceVar(&emails, "emails", nil, "Comma-separated viewer email addresses")
	cmd.MarkFlagRequired("emails")
	return cmd
}

func readViewers(cmd *cobra.Command, path string) ([]sdk.ChannelViewerInviteParamsUser, error) {
	input := cmd.InOrStdin()
	if path != "-" {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open --viewers-file: %w", err)
		}
		defer file.Close()
		input = file
	}
	const maxBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(input, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read --viewers-file: %w", err)
	}
	if len(data) > maxBytes {
		return nil, fmt.Errorf("--viewers-file exceeds 1 MiB")
	}
	var entries []struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if json.Unmarshal(data, &entries) != nil {
		return nil, fmt.Errorf("--viewers-file must contain a JSON array of {email,name} objects")
	}
	if len(entries) < 1 || len(entries) > 200 {
		return nil, fmt.Errorf("--viewers-file must contain 1–200 viewers")
	}
	users := make([]sdk.ChannelViewerInviteParamsUser, 0, len(entries))
	for i, entry := range entries {
		if !validEmail(entry.Email) {
			return nil, fmt.Errorf("viewer %d must have a bare email address", i+1)
		}
		if strings.TrimSpace(entry.Name) == "" || utf8.RuneCountInString(entry.Name) > 100 {
			return nil, fmt.Errorf("viewer %d must have a name of 1–100 characters", i+1)
		}
		users = append(users, sdk.ChannelViewerInviteParamsUser{Email: sdk.F(entry.Email), Name: sdk.F(entry.Name)})
	}
	return users, nil
}

func validEmail(value string) bool {
	addr, err := mail.ParseAddress(value)
	return err == nil && addr.Name == "" && addr.Address == value
}
