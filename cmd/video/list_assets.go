package video

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var listAssetsCmd = &cobra.Command{
	Use:   "list-assets",
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

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(listAssetsCmd)
	listAssetsCmd.Flags().String("workspace-id", "", "Workspace ID to list assets from")
	listAssetsCmd.MarkFlagRequired("workspace-id")
	listAssetsCmd.Flags().String("status", "", "Filter by asset status")
	listAssetsCmd.Flags().String("tag", "", "Filter by asset tag")
	listAssetsCmd.Flags().String("title", "", "Filter by asset title")
	listAssetsCmd.Flags().String("folder", "", "Filter by asset folder")
	listAssetsCmd.Flags().String("offset", "", "Offset for pagination")
	listAssetsCmd.Flags().String("size", "", "Page size for pagination")
	listAssetsCmd.Flags().String("playlist-id", "", "Filter by playlist ID")
	listAssetsCmd.Flags().String("sort-by", "", "Field to sort by")
	listAssetsCmd.Flags().String("order-by", "", "Order to sort by")
	listAssetsCmd.Flags().String("type", "", "Type of asset to filter")
}
