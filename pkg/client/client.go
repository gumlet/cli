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
}

type Factory func() (*Client, error)

func New(apiKey string, extra ...option.RequestOption) *Client {
	opts := []option.RequestOption{
		option.WithEnvironmentProduction(),
		option.WithAPIKey(apiKey),
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
	}
}
