package client

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gumlet/pkg/ui"

	sdk "github.com/gumlet/golang-sdk"
	"github.com/gumlet/golang-sdk/option"
	"github.com/spf13/viper"
)

type Client struct {
	ImageSources    *sdk.ImageSourceService
	VideoAssets     *sdk.VideoAssetService
	VideoPlaylists  *sdk.VideoPlaylistService
	VideoWorkspaces *sdk.VideoWorkspaceService
}

func NewClient() (*Client, error) {
	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		return nil, fmt.Errorf("API key is not set — run 'gumlet login' to get started")
	}
	opts := []option.RequestOption{
		option.WithEnvironmentProduction(),
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(30 * time.Second),
		// Keep the previous one-attempt behavior, especially for asset creation.
		option.WithMaxRetries(0),
	}
	// Construct the SDK services explicitly so inherited SDK environment
	// variables can't redirect a saved CLI credential or inject headers.
	return &Client{
		ImageSources:    sdk.NewImageSourceService(opts...),
		VideoAssets:     sdk.NewVideoAssetService(opts...),
		VideoPlaylists:  sdk.NewVideoPlaylistService(opts...),
		VideoWorkspaces: sdk.NewVideoWorkspaceService(opts...),
	}, nil
}

type progressReader struct {
	r         io.Reader
	total     int64
	read      int64
	lastPrint int64
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	p.read += int64(n)
	if p.total > 0 && (p.read-p.lastPrint >= 256*1024 || err == io.EOF) {
		ui.ProgressBar("Uploading", p.read, p.total)
		p.lastPrint = p.read
	}
	return n, err
}

// PutFile streams a local file to a pre-signed URL (e.g. S3) using a plain PUT.
// No Gumlet auth header is added — the URL is self-authorising.
func PutFile(ctx context.Context, uploadURL, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	pr := &progressReader{r: f, total: info.Size()}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, pr)
	if err != nil {
		return err
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	ui.FinishProgress()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}
