package image

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var deleteSourceCmd = &cobra.Command{
	Use:   "delete-source",
	Short: "Delete an image source",
	Run: func(cmd *cobra.Command, args []string) {
		sourceID, _ := cmd.Flags().GetString("source-id")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/image/sources/%s", sourceID)
		resp, err := apiClient.Delete(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(deleteSourceCmd)
	deleteSourceCmd.Flags().String("source-id", "", "ID of the image source to delete")
	deleteSourceCmd.MarkFlagRequired("source-id")
}
