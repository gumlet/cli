package image

import (
	"fmt"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/spf13/cobra"
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge image cache for a source",
	Run: func(cmd *cobra.Command, args []string) {
		subdomain, _ := cmd.Flags().GetString("subdomain")
		urls, _ := cmd.Flags().GetStringSlice("urls")
		output, _ := cmd.Root().PersistentFlags().GetString("output")

		body := map[string]interface{}{
			"paths": urls,
		}

		apiClient, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/purge/%s", subdomain)
		resp, err := apiClient.Post(path, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		printer.Print(resp, output)
	},
}

func init() {
	Cmd.AddCommand(purgeCmd)
	purgeCmd.Flags().String("subdomain", "", "Subdomain to purge cache for")
	purgeCmd.MarkFlagRequired("subdomain")
	purgeCmd.Flags().StringSlice("urls", []string{}, "URLs to purge")
}
