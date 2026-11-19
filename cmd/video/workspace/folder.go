package workspace

import (
	"errors"
	"strings"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

var folderRows = printer.Table{Columns: []string{"name", "id", "parent_id", "asset_count", "subdirectory_count"}}

func newFolderCommand(newClient client.Factory) *cobra.Command {
	group := &cobra.Command{Use: "folder", Short: "Manage video workspace folders"}
	var workspaceID string
	group.PersistentFlags().StringVar(&workspaceID, "workspace-id", "", "ID of the workspace")
	group.MarkPersistentFlagRequired("workspace-id")

	create := &cobra.Command{
		Use: "create", Short: "Create a folder", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			parent, _ := cmd.Flags().GetString("parent-id")
			if strings.TrimSpace(name) == "" {
				return errors.New("--name must not be empty")
			}
			if cmd.Flags().Changed("parent-id") && strings.TrimSpace(parent) == "" {
				return errors.New("--parent-id must not be empty")
			}
			params := sdk.FolderNewParams{Name: sdk.F(name)}
			if cmd.Flags().Changed("parent-id") {
				params.ParentID = sdk.F(parent)
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.Folders.New(cmd.Context(), workspaceID, params, option.WithResponseBodyInto(&raw))
			table := folderRows
			table.Result = "Folder created."
			return printFolder(cmd, raw, table, err)
		},
	}
	create.Flags().String("name", "", "Folder name")
	create.MarkFlagRequired("name")
	create.Flags().String("parent-id", "", "Parent folder ID; omit to create a root folder")

	list := &cobra.Command{
		Use: "list", Short: "List folders", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			parent, _ := cmd.Flags().GetString("parent-id")
			params := sdk.FolderListParams{}
			if cmd.Flags().Changed("parent-id") {
				if strings.TrimSpace(parent) == "" {
					return errors.New("--parent-id must not be empty")
				}
				params.ParentID = sdk.F(parent)
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.Folders.List(cmd.Context(), workspaceID, params, option.WithResponseBodyInto(&raw))
			return printFolder(cmd, raw, folderRows, err)
		},
	}
	list.Flags().String("parent-id", "", `Filter by parent folder ID; use "null" for root folders`)

	get := newFolderByIDCommand("get", "Get a folder", func(cmd *cobra.Command, folderID string) error {
		api, err := newClient()
		if err != nil {
			return err
		}
		var raw []byte
		_, err = api.Folders.Get(cmd.Context(), workspaceID, folderID, option.WithResponseBodyInto(&raw))
		return printFolder(cmd, raw, printer.Table{
			Columns: []string{"name", "id", "parent_id", "path_names", "asset_count", "subdirectory_count", "updated_at"},
			Fields: map[string]printer.Field{
				"path_names": {Label: "Folder path", Format: printer.JoinStrings(" / ")},
				"updated_at": {Label: "Updated", Format: printer.Date(printer.UTCDateTime, nil)},
			},
		}, err)
	})
	deleteFolder := newFolderByIDCommand("delete", "Delete a folder", func(cmd *cobra.Command, folderID string) error {
		api, err := newClient()
		if err != nil {
			return err
		}
		var raw []byte
		_, err = api.Folders.Delete(cmd.Context(), workspaceID, folderID, option.WithResponseBodyInto(&raw))
		return printFolder(cmd, raw, printer.Table{Result: "Folder deleted."}, err)
	})
	update := newFolderByIDCommand("update", "Rename a folder or move assets into it", func(cmd *cobra.Command, folderID string) error {
		flags := cmd.Flags()
		name, _ := flags.GetString("name")
		parent, _ := flags.GetString("parent-id")
		root, _ := flags.GetBool("root")
		ids, _ := flags.GetStringSlice("asset-ids")

		if !flags.Changed("name") && !flags.Changed("asset-ids") {
			return errors.New("specify --name or --asset-ids")
		}
		if flags.Changed("name") && strings.TrimSpace(name) == "" {
			return errors.New("--name must not be empty")
		}
		if flags.Changed("parent-id") && strings.TrimSpace(parent) == "" {
			return errors.New("--parent-id must not be empty")
		}
		if root && flags.Changed("parent-id") {
			return errors.New("--root and --parent-id cannot be combined")
		}
		// Only the name-bearing update variant accepts a null parent.
		if root && !flags.Changed("name") {
			return errors.New("--root requires --name (use the current name to keep it)")
		}
		body := sdk.FolderUpdateParamsBody{}
		if flags.Changed("name") {
			body.Name = sdk.F(name)
		}
		if flags.Changed("asset-ids") {
			if err := validateFolderAssetIDs(ids); err != nil {
				return err
			}
			body.AssetIDs = sdk.F[any](ids)
		}
		if root {
			body.ParentID = sdk.Null[string]()
		} else if flags.Changed("parent-id") {
			body.ParentID = sdk.F(parent)
		}
		api, err := newClient()
		if err != nil {
			return err
		}
		var raw []byte
		_, err = api.Folders.Update(cmd.Context(), workspaceID, folderID, sdk.FolderUpdateParams{Body: body}, option.WithResponseBodyInto(&raw))
		return printFolder(cmd, raw, printer.Table{
			Result:  "Folder updated.",
			Columns: []string{"name", "id", "parent_id", "movedCount"},
			Fields:  map[string]printer.Field{"movedCount": {Label: "Assets moved"}},
		}, err)
	})
	update.Flags().String("name", "", "New folder name")
	update.Flags().String("parent-id", "", "New parent folder ID")
	update.Flags().Bool("root", false, "Move the folder to the root level (requires --name)")
	update.Flags().StringSlice("asset-ids", nil, "Asset IDs to move into the folder")

	remove := &cobra.Command{
		Use: "remove-assets", Short: "Remove assets from their folders without deleting them", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, _ := cmd.Flags().GetStringSlice("asset-ids")
			if err := validateFolderAssetIDs(ids); err != nil {
				return err
			}
			api, err := newClient()
			if err != nil {
				return err
			}
			var raw []byte
			_, err = api.Folders.RemoveAssets(cmd.Context(), workspaceID, sdk.FolderRemoveAssetsParams{AssetIDs: sdk.F(ids)}, option.WithResponseBodyInto(&raw))
			return printFolder(cmd, raw, printer.Table{
				Result:  "Assets removed from folders.",
				Columns: []string{"removedCount"},
				Fields:  map[string]printer.Field{"removedCount": {Label: "Assets removed"}},
			}, err)
		},
	}
	remove.Flags().StringSlice("asset-ids", nil, "Asset IDs to remove from folders")
	remove.MarkFlagRequired("asset-ids")
	for _, sub := range []*cobra.Command{create, list, get, update, deleteFolder, remove} {
		sub.PreRunE = func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("workspace-id") && strings.TrimSpace(workspaceID) == "" {
				return errors.New("--workspace-id must not be empty")
			}
			return nil
		}
	}
	group.AddCommand(create, list, get, update, deleteFolder, remove)
	return group
}

func newFolderByIDCommand(use, short string, run func(*cobra.Command, string) error) *cobra.Command {
	var folderID string
	cmd := &cobra.Command{
		Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(folderID) == "" {
				return errors.New("--folder-id must not be empty")
			}
			return run(cmd, folderID)
		},
	}
	cmd.Flags().StringVar(&folderID, "folder-id", "", "ID of the folder")
	cmd.MarkFlagRequired("folder-id")
	return cmd
}

func validateFolderAssetIDs(ids []string) error {
	if len(ids) == 0 {
		return errors.New("--asset-ids must include at least one asset ID")
	}
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return errors.New("--asset-ids must not contain empty IDs")
		}
	}
	return nil
}

func printFolder(cmd *cobra.Command, raw []byte, table printer.Table, err error) error {
	if err != nil {
		return err
	}
	return printer.Print(cmd, raw, table)
}
