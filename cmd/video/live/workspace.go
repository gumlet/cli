package live

import (
	"context"
	"fmt"
	"strings"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var workspaceTable = printer.Table{Columns: []string{"id", "name", "video_workspace_id"}}

func newWorkspaceCommand(newClient client.Factory) *cobra.Command {
	group := &cobra.Command{Use: "workspace", Short: "Manage live streaming workspaces"}
	list := leaf("list", "List live workspaces", func(cmd *cobra.Command) error {
		table := workspaceTable
		table.RowsKey = "all_live_sources"
		return runRequest(cmd, newClient, table, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveWorkspaces.List(ctx, opt)
			return err
		})
	})
	var createName string
	create := leaf("create", "Create a live workspace", func(cmd *cobra.Command) error {
		if strings.TrimSpace(createName) == "" {
			return fmt.Errorf("--name must not be empty")
		}
		return runRequest(cmd, newClient, workspaceTable, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveWorkspaces.New(ctx, sdk.LiveStreamWorkspaceNewParams{Name: sdk.F(createName)}, opt)
			return err
		})
	})
	requireString(create, &createName, "name", "Name of the live workspace")

	var updateID, name, vodWorkspace string
	update := leaf("update", "Update a live workspace", func(cmd *cobra.Command) error {
		if strings.TrimSpace(updateID) == "" {
			return fmt.Errorf("--live-workspace-id must not be empty")
		}
		params := sdk.LiveStreamWorkspaceUpdateParams{}
		if cmd.Flags().Changed("name") {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("--name must not be empty")
			}
			params.Name = sdk.F(name)
		}
		if cmd.Flags().Changed("video-workspace-id") {
			if strings.TrimSpace(vodWorkspace) == "" {
				return fmt.Errorf("--video-workspace-id must not be empty")
			}
			params.VideoSourceID = sdk.F(vodWorkspace)
		}
		return runRequest(cmd, newClient, workspaceTable, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveWorkspaces.Update(ctx, updateID, params, opt)
			return err
		})
	})
	requireString(update, &updateID, "live-workspace-id", "ID of the live workspace")
	update.Flags().StringVar(&name, "name", "", "New name")
	update.Flags().StringVar(&vodWorkspace, "video-workspace-id", "", "ID of the associated video-on-demand workspace")
	update.MarkFlagsOneRequired("name", "video-workspace-id")

	var deleteID string
	remove := leaf("delete", "Delete a live workspace", func(cmd *cobra.Command) error {
		if strings.TrimSpace(deleteID) == "" {
			return fmt.Errorf("--live-workspace-id must not be empty")
		}
		return runRequest(cmd, newClient, printer.Table{}, func(ctx context.Context, api *client.Client, opt option.RequestOption) error {
			_, err := api.LiveWorkspaces.Delete(ctx, deleteID, opt)
			return err
		})
	})
	requireString(remove, &deleteID, "live-workspace-id", "ID of the live workspace")
	group.AddCommand(list, create, update, remove)
	return group
}
