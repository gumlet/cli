package asset

import (
	"time"

	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newGetCommand(newClient client.Factory) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a video asset",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assetID, _ := cmd.Flags().GetString("asset-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoAssets.GetDetails(cmd.Context(), assetID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			return printer.Print(cmd, resp, printer.Table{
				Columns: []string{
					"asset_id", "input.title", "status", "workspace_id", "input.description",
					"input.duration", "tag", "created_at", "updated_at", "output.playback_url", "output.thumbnail_url",
				},
				Fields: map[string]printer.Field{
					"input.duration":       {Label: "Duration (s)"},
					"tag":                  {Label: "Tags", Format: printer.JoinStrings(", ")},
					"created_at":           {Label: "Created", Format: printer.Date(printer.UTCDateTime, time.UnixMilli)},
					"updated_at":           {Label: "Updated", Format: printer.Date(printer.UTCDateTime, time.UnixMilli)},
					"output.playback_url":  {Format: printer.PublicLink},
					"output.thumbnail_url": {Format: printer.PublicLink},
				},
			})
		},
	}
	getCmd.Flags().String("asset-id", "", "ID of the asset")
	getCmd.MarkFlagRequired("asset-id")
	return getCmd
}
