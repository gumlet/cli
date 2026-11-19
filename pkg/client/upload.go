package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type uploadReader struct {
	io.Reader
	read     int64
	progress func(int64)
}

func (r *uploadReader) Read(buf []byte) (int, error) {
	n, err := r.Reader.Read(buf)
	if n > 0 {
		r.read += int64(n)
		r.progress(r.read)
	}
	return n, err
}

// Upload streams an already-open file to a signed URL without API credentials
// or redirects. Errors must not print the signed URL or the response body.
func Upload(ctx context.Context, signedURL string, input io.Reader, size int64, contentType string, progress func(int64)) error {
	var body io.Reader = io.LimitReader(input, size)
	if progress != nil {
		body = &uploadReader{Reader: body, progress: progress}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, body)
	if err != nil || req.URL.Host == "" || req.URL.User != nil || (req.URL.Scheme != "https" && req.URL.Scheme != "http") {
		return errors.New("invalid signed upload URL")
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", contentType)

	transfer := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := transfer.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("could not transfer file to storage")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("storage returned HTTP %d", resp.StatusCode)
	}
	return nil
}
