package asset

import (
	"fmt"
	"strings"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newBulkDeleteCommand(newClient client.Factory) *cobra.Command {
	return newBulkCommand("bulk-delete", "Delete multiple video assets",
		func(cmd *cobra.Command, workspaceID string, ids []string) error {
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.VideoAssets.DeleteMany(cmd.Context(), sdk.VideoAssetDeleteManyParams{
				AssetList: sdk.F(ids), WorkspaceID: sdk.F(workspaceID),
			}, option.WithResponseBodyInto(&raw))
			if err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{})
		})
}

func newBulkTagCommand(newClient client.Factory) *cobra.Command {
	var add, remove []string
	cmd := newBulkCommand("bulk-tag", "Add or remove tags on multiple video assets",
		func(cmd *cobra.Command, workspaceID string, ids []string) error {
			if len(add) == 0 && len(remove) == 0 {
				return fmt.Errorf("provide at least one tag with --add-tags or --remove-tags")
			}
			if err := rejectBlankEntries("add-tags", add); err != nil {
				return err
			}
			if err := rejectBlankEntries("remove-tags", remove); err != nil {
				return err
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.VideoAssets.TagMany(cmd.Context(), sdk.VideoAssetTagManyParams{
				AssetList: sdk.F(ids), WorkspaceID: sdk.F(workspaceID),
				AddTags: sdk.F(append([]string{}, add...)), RemoveTags: sdk.F(append([]string{}, remove...)),
			}, option.WithResponseBodyInto(&raw))
			if err != nil {
				return err
			}
			return printer.Print(cmd, raw, printer.Table{})
		})
	cmd.Flags().StringSliceVar(&add, "add-tags", nil, "Tags to add")
	cmd.Flags().StringSliceVar(&remove, "remove-tags", nil, "Tags to remove")
	return cmd
}

func newBulkCommand(use, short string, run func(*cobra.Command, string, []string) error) *cobra.Command {
	var workspaceID string
	var ids []string
	cmd := &cobra.Command{
		Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(workspaceID) == "" {
				return fmt.Errorf("--workspace-id must not be empty")
			}
			if len(ids) == 0 {
				return fmt.Errorf("--asset-ids must include at least one asset ID")
			}
			if err := rejectBlankEntries("asset-ids", ids); err != nil {
				return err
			}
			return run(cmd, workspaceID, ids)
		},
	}
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "ID of the workspace")
	cmd.Flags().StringSliceVar(&ids, "asset-ids", nil, "Asset IDs to modify")
	cmd.MarkFlagRequired("workspace-id")
	cmd.MarkFlagRequired("asset-ids")
	return cmd
}

func rejectBlankEntries(flag string, values []string) error {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("--%s must not contain empty entries", flag)
		}
	}
	return nil
}
