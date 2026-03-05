package asset

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List assets in a video workspace",
	Run: func(cmd *cobra.Command, args []string) {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/video/assets/list/%s", workspaceID)

		queryParams := make(map[string]string)
		cmd.Flags().Visit(func(f *pflag.Flag) {
			if f.Name != "workspace-id" {
				queryParams[f.Name] = f.Value.String()
			}
		})

		resp, err := apiClient.Get(path, queryParams)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output, "asset_id", "input.title", "status", "created_at", "tag")
	},
}

func init() {
	Cmd.AddCommand(listCmd)
	listCmd.Flags().String("workspace-id", "", "Workspace ID to list assets from")
	listCmd.MarkFlagRequired("workspace-id")
	listCmd.Flags().String("status", "", "Filter by asset status")
	listCmd.Flags().String("tag", "", "Filter by asset tag")
	listCmd.Flags().String("title", "", "Filter by asset title")
	listCmd.Flags().String("folder", "", "Filter by asset folder")
	listCmd.Flags().String("offset", "", "Offset for pagination")
	listCmd.Flags().String("size", "", "Page size for pagination")
	listCmd.Flags().String("playlist-id", "", "Filter by playlist ID")
	listCmd.Flags().String("sort-by", "", "Field to sort by")
	listCmd.Flags().String("order-by", "", "Order to sort by")
	listCmd.Flags().String("type", "", "Type of asset to filter")
}
