package workspace

import (
	"gumlet/pkg/client"
	"gumlet/pkg/printer"

	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/cobra"
)

func newGetCommand(newClient client.Factory) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a video workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID, _ := cmd.Flags().GetString("workspace-id")

			apiClient, err := newClient()
			if err != nil {
				return err
			}

			var resp []byte
			_, err = apiClient.VideoWorkspaces.Get(cmd.Context(), workspaceID, option.WithResponseBodyInto(&resp))
			if err != nil {
				return err
			}

			timestamp := printer.Date(printer.UTCDateTime, nil)
			return printer.Print(cmd, resp, printer.Table{
				Columns: []string{"id", "name", "type", "created_at", "updated_at"},
				Summary: []string{
					"default_profile_id", "video_protection.signed_url",
					"channel_settings.title", "channel_settings.privacy_type",
				},
				Fields: map[string]printer.Field{
					"created_at":                    {Label: "Created", Format: timestamp},
					"updated_at":                    {Label: "Updated", Format: timestamp},
					"default_profile_id":            {Label: "Default profile ID"},
					"video_protection.signed_url":   {Label: "Signed URLs"},
					"channel_settings.title":        {Label: "Channel title"},
					"channel_settings.privacy_type": {Label: "Channel privacy"},
				},
			})
		},
	}
	getCmd.Flags().String("workspace-id", "", "ID of the workspace")
	getCmd.MarkFlagRequired("workspace-id")
	return getCmd
}
