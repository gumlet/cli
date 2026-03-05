package purge

import (
	"fmt"

	"gumlet/pkg/client"

	"github.com/spf13/cobra"
)

var purgeImageCacheCmd = &cobra.Command{
	Use:   "image-cache",
	Short: "Purge image cache",
	Run: func(cmd *cobra.Command, args []string) {
		subdomain, _ := cmd.Flags().GetString("subdomain")
		urls, _ := cmd.Flags().GetStringSlice("urls")

		body := map[string]interface{}{
			"urls": urls,
		}

		client, err := client.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := fmt.Sprintf("/purge/%s", subdomain)
		resp, err := client.Post(path, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(resp))
	},
}

func init() {
	Cmd.AddCommand(purgeImageCacheCmd)
	purgeImageCacheCmd.Flags().String("subdomain", "", "Subdomain to purge cache from")
	purgeImageCacheCmd.MarkFlagRequired("subdomain")
	purgeImageCacheCmd.Flags().StringSlice("urls", []string{}, "URLs to purge")
}
