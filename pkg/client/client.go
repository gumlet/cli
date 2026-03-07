package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const baseURL = "https://api.gumlet.com/v1"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient() (*Client, error) {
	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		return nil, fmt.Errorf("API key is not set. Please run 'gumlet login' to configure your API key")
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	url := baseURL + path
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return req, nil
}

func (c *Client) Do(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	return body, nil
}

func (c *Client) Get(path string, queryParams map[string]string) ([]byte, error) {
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range queryParams {
		if v != "" {
			q.Add(k, v)
		}
	}
	req.URL.RawQuery = q.Encode()

	return c.Do(req)
}

func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Delete(path string) ([]byte, error) {
	req, err := c.newRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) DeleteWithBody(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("DELETE", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	req, err := c.newRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// progressReader wraps an io.Reader and prints upload progress to stdout.
type progressReader struct {
	r         io.Reader
	total     int64
	read      int64
	lastPrint int64
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	p.read += int64(n)
	// Print every ~1 MB or when done
	if p.total > 0 && (p.read-p.lastPrint >= 1024*1024 || err == io.EOF) {
		pct := float64(p.read) / float64(p.total) * 100
		fmt.Fprintf(os.Stderr, "\rUploading... %.1f%% (%s / %s)",
			pct, formatBytes(p.read), formatBytes(p.total))
		p.lastPrint = p.read
	}
	return n, err
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// PutFile streams a local file to a pre-signed URL (e.g. S3) using a plain PUT.
// No Gumlet auth header is added — the URL is self-authorising.
func (c *Client) PutFile(uploadURL, filePath string) error {
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
	req, err := http.NewRequest("PUT", uploadURL, pr)
	if err != nil {
		return err
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	fmt.Fprintln(os.Stderr) // newline after progress line
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
