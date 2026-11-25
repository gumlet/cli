package client

import (
	"time"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
)

type Client struct {
	ImageSources    *sdk.ImageSourceService
	VideoAssets     *sdk.VideoAssetService
	VideoPlaylists  *sdk.VideoPlaylistService
	VideoWorkspaces *sdk.VideoWorkspaceService
	Folders         *sdk.FolderService
	RecycleBin      *sdk.RecycleBinService
	UserData        *sdk.UserDataService
	Billing         *sdk.BillingService
	ChannelViewers   *sdk.ChannelViewerService
	LiveWorkspaces   *sdk.LiveStreamWorkspaceService
	LiveAssets       *sdk.LiveStreamAssetService
	OrganizationData *sdk.OrganizationDataService
	AuditLogs        *sdk.AuditLogService
}

type Factory func() (*Client, error)

func New(apiKey, version string, extra ...option.RequestOption) *Client {
	if version == "" {
		version = "dev"
	}
	opts := []option.RequestOption{
		option.WithEnvironmentProduction(),
		option.WithAPIKey(apiKey),
		option.WithHeader("User-Agent", "gumlet-cli/"+version),
		option.WithRequestTimeout(30 * time.Second),
		// Keep the previous one-attempt behavior, especially for asset creation.
		option.WithMaxRetries(0),
	}
	opts = append(opts, extra...)
	// Construct the SDK services explicitly so inherited SDK environment
	// variables can't redirect a saved CLI credential or inject headers.
	return &Client{
		ImageSources:    sdk.NewImageSourceService(opts...),
		VideoAssets:     sdk.NewVideoAssetService(opts...),
		VideoPlaylists:  sdk.NewVideoPlaylistService(opts...),
		VideoWorkspaces: sdk.NewVideoWorkspaceService(opts...),
		Folders:         sdk.NewFolderService(opts...),
		RecycleBin:      sdk.NewRecycleBinService(opts...),
		UserData:        sdk.NewUserDataService(opts...),
		Billing:         sdk.NewBillingService(opts...),
		ChannelViewers:   sdk.NewChannelViewerService(opts...),
		LiveWorkspaces:   sdk.NewLiveStreamWorkspaceService(opts...),
		LiveAssets:       sdk.NewLiveStreamAssetService(opts...),
		OrganizationData: sdk.NewOrganizationDataService(opts...),
		AuditLogs:        sdk.NewAuditLogService(opts...),
	}
}
