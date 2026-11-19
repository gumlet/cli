package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gumlet/cmd"
	"gumlet/pkg/client"
	"gumlet/pkg/credentials"

	"github.com/gumlet/golang-sdk/option"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func runCLI(t *testing.T, baseURL, input string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	target, err := url.Parse(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	transport := &http.Transport{}
	defer transport.CloseIdleConnections()
	httpClient := &http.Client{Transport: testTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme != "https" || r.URL.Host != "api.gumlet.com" {
			return nil, fmt.Errorf("unexpected API origin: %s://%s", r.URL.Scheme, r.URL.Host)
		}
		r = r.Clone(r.Context())
		r.URL.Scheme, r.URL.Host, r.Host = target.Scheme, target.Host, target.Host
		return transport.RoundTrip(r)
	})}

	store := credentials.Store{Path: filepath.Join(t.TempDir(), ".gumlet.yaml")}
	if _, err := store.Save("test-key"); err != nil {
		t.Fatal(err)
	}
	t.Setenv(credentials.EnvKey, "")
	t.Setenv("GUMLET_BASE_URL", baseURL+"/unexpected/")
	t.Setenv("GUMLET_CUSTOM_HEADERS", "X-Unexpected: injected")
	t.Setenv("API_KEY", "wrong-key")

	root := cmd.NewRootCommand(cmd.Options{NewClient: func() (*client.Client, error) {
		key, err := store.APIKey()
		if err != nil {
			return nil, err
		}
		return client.New(key, option.WithHTTPClient(httpClient)), nil
	}})
	var out, diagnostics bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&diagnostics)
	root.SetIn(strings.NewReader(input))
	root.SetArgs(args)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = root.ExecuteContext(ctx)
	return out.String(), diagnostics.String(), err
}

func checkJSON(t *testing.T, actual []byte, expected string) {
	t.Helper()
	var got, want any
	if json.Unmarshal(actual, &got) != nil || json.Unmarshal([]byte(expected), &want) != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("JSON = %s; want %s", actual, expected)
	}
}

func TestSDKRequests(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, query, body, response, output string
		args                                              []string
		status                                            int
	}{
		{"asset list", "GET", "/v1/video/assets/list/ws", "folder_id=&orderBy=asc&page_size=10&playlist_id=pl&sortBy=created_at&status=ready", "", `{"all_assets":[{"asset_id":"one","input":{"title":"Example"},"extra":1}],"total_asset_count":1}`, `{"all_assets":[{"asset_id":"one","input":{"title":"Example"},"extra":1}],"total_asset_count":1}`, []string{"video", "asset", "list", "--workspace-id", "ws", "--status", "ready", "--sort-by", "created_at", "--order-by", "asc", "--page-size", "10", "--folder-id=", "--playlist-id", "pl"}, 200},
		{"workspace create", "POST", "/v1/video/workspaces", "", `{"name":"Example"}`, `{"id":"ws"}`, `{"id":"ws"}`, []string{"video", "workspace", "create", "--name", "Example"}, 200},
		{"playlist create", "POST", "/v1/video/playlist", "", `{"workspace_id":"ws","title":"Example"}`, `{"id":"pl"}`, `{"id":"pl"}`, []string{"video", "playlist", "create", "--workspace-id", "ws", "--title", "Example"}, 200},
		{"playlist list", "GET", "/v1/video/playlist", "workspace_id=ws", "", `[{"id":"pl"}]`, `[{"id":"pl"}]`, []string{"video", "playlist", "list", "--workspace-id", "ws"}, 200},
		{"playlist update", "POST", "/v1/video/playlist/pl", "", `{"channel_visibility":false}`, `{"id":"pl"}`, `{"id":"pl"}`, []string{"video", "playlist", "update", "--playlist-id", "pl", "--channel-visibility=false"}, 200},
		{"playlist delete", "DELETE", "/v1/video/playlist/pl", "", "", `{"deleted":true}`, `{"deleted":true}`, []string{"video", "playlist", "delete", "--playlist-id", "pl"}, 200},
		{"empty delete", "DELETE", "/v1/video/assets/asset", "", "", "", "", []string{"video", "asset", "delete", "--asset-id", "asset"}, 204},
		{"legacy source config", "POST", "/v1/image/sources", "", `{"namespace":"test","type":"amazon","amazon":{"bucket_name":"example"}}`, `{"id":"src"}`, `{"id":"src"}`, []string{"image", "source", "add", "--namespace", "test", "--type", "amazon", "--config", `{"bucket_name":"example"}`}, 200},
		{"legacy source update", "POST", "/v1/image/sources/src", "", `{"type":"amazon","amazon":{"bucket_name":"example"}}`, `{"id":"src"}`, `{"id":"src"}`, []string{"image", "source", "update", "--source-id", "src", "--type", "amazon", "--config", `{"bucket_name":"example"}`}, 200},
		{"full purge", "POST", "/v1/purge/example", "", `{"paths":[]}`, "", "", []string{"image", "purge", "--subdomain", "example"}, 204},
		{"failed mutation", "POST", "/v1/video/workspaces", "", `{"name":"Example"}`, `{"message":"unavailable"}`, "", []string{"video", "workspace", "create", "--name", "Example"}, 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Method != tc.method || r.URL.Path != tc.path || r.URL.RawQuery != tc.query {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Errorf("unexpected Authorization: %q", got)
				}
				if got := r.Header.Get("X-Unexpected"); got != "" {
					t.Errorf("unexpected header from SDK environment: %q", got)
				}
				if tc.body != "" {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
					}
					checkJSON(t, body, tc.body)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					_, _ = io.WriteString(w, tc.response)
				}
			}))
			defer server.Close()

			out, stderr, err := runCLI(t, server.URL, "", tc.args...)
			if stderr != "" {
				t.Errorf("unexpected diagnostics: %s", stderr)
			}
			if tc.status >= 400 {
				if err == nil || !strings.Contains(err.Error(), "503") || out != "" {
					t.Fatalf("expected API failure, got %v: %s", err, out)
				}
			} else if err != nil {
				t.Fatalf("command failed: %v: %s", err, out)
			} else if tc.output != "" {
				checkJSON(t, []byte(out), tc.output)
			} else if out != "" {
				t.Errorf("empty response printed %q", out)
			}
			if got := requests.Load(); got != 1 {
				t.Errorf("request count = %d; want 1", got)
			}
		})
	}
}

func TestAssetCursorTable(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("next_page_token") != "cursor+1" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ignored":[{"asset_id":"wrong"}],"all_assets":[{"asset_id":"asset","input":{"title":"Example"}}]}`)
	}))
	defer api.Close()
	out, stderr, err := runCLI(t, api.URL, "", "video", "asset", "list", "--workspace-id", "ws", "--next-page-token", "cursor+1", "-o", "table")
	if err != nil || stderr != "" || !strings.Contains(out, "Example") || strings.Contains(out, "wrong") {
		t.Fatalf("stdout = %q, stderr = %q, err = %v", out, stderr, err)
	}
}

func TestSourceConfigInput(t *testing.T) {
	const config = `{"bucket_name":"example","number":9007199254740993}`
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || string(body["amazon"]) != config {
			t.Errorf("unexpected config: %s (%v)", body["amazon"], err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer api.Close()
	path := filepath.Join(t.TempDir(), "source.json")
	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	for _, filename := range []string{"-", path} {
		_, _, err := runCLI(t, api.URL, config, "image", "source", "add", "--namespace", "test", "--type", "amazon", "--config-file", filename)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestValidationBeforeClientCreation(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.mp4")
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"video", "asset", "upload", "--workspace-id", "ws", "--file", missing}, "inspect video"},
		{[]string{"video", "asset", "list", "--workspace-id", "ws", "--page-size", "101"}, "between 1 and 100"},
		{[]string{"video", "asset", "list", "--workspace-id", "ws", "--next-page-token", "token", "--title", "test"}, "no title filter"},
		{[]string{"video", "workspace", "list", "-o", "csv"}, "invalid --output"},
		{[]string{"image", "source", "add", "--namespace", "test", "--type", "amazon", "--config", "null"}, "valid JSON object"},
		{[]string{"image", "source", "update", "--source-id", "src", "--config", "{}"}, "type"},
	} {
		root := cmd.NewRootCommand(cmd.Options{NewClient: func() (*client.Client, error) {
			t.Fatal("created API client for invalid input")
			return nil, nil
		}})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs(tc.args)
		if err := root.Execute(); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v: got %v, want %q", tc.args, err, tc.want)
		}
	}
}

func TestLoginPrintsKeyDashboard(t *testing.T) {
	_, stderr, err := runCLI(t, "", "", "login")
	if err == nil || !strings.Contains(stderr, "https://dash.gumlet.com/developer/api-keys") {
		t.Fatalf("login diagnostics = %q, err = %v", stderr, err)
	}
}

func TestSDKUpload(t *testing.T) {
	video := "sample video bytes"
	file := filepath.Join(t.TempDir(), "sample.mp4")
	if err := os.WriteFile(file, []byte(video), 0600); err != nil {
		t.Fatal(err)
	}
	var uploads atomic.Int32
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploads.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != video || r.Method != "PUT" || r.URL.RawQuery != "signature=test" || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected signed upload: %s %s, authorization present=%t, body=%q, err=%v", r.Method, r.URL, r.Header.Get("Authorization") != "", body, err)
		}
	}))
	defer storage.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/video/assets/upload" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected session request: %s %s", r.Method, r.URL)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		checkJSON(t, body, `{"workspace_id":"ws","format":"ABR","title":"sample.mp4","tag":["one","two"]}`)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"asset_id": "asset", "upload_url": storage.URL + "/file?signature=test"})
	}))
	defer api.Close()

	out, stderr, err := runCLI(t, api.URL, "", "video", "asset", "upload", "--file", file, "--workspace-id", "ws", "--tag", "one,two")
	if err != nil || uploads.Load() != 1 || !strings.Contains(stderr, "Uploading video") {
		t.Fatalf("upload count=%d, err=%v, output=%s", uploads.Load(), err, out)
	}
	checkJSON(t, []byte(out), `{"asset_id":"asset","uploaded":true}`)
}

func TestUploadFailureReportsAssetNotSignature(t *testing.T) {
	file := filepath.Join(t.TempDir(), "sample.mp4")
	if err := os.WriteFile(file, []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}
	var unexpectedRedirect atomic.Bool
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirected" {
			unexpectedRedirect.Store(true)
		}
		http.Redirect(w, r, "/redirected", http.StatusSeeOther)
	}))
	defer storage.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"asset_id": "new-asset", "upload_url": storage.URL + "/file?signature=secret-test"})
	}))
	defer api.Close()

	out, stderr, err := runCLI(t, api.URL, "", "video", "asset", "upload", "--file", file, "--workspace-id", "ws")
	if err == nil || !strings.Contains(err.Error(), "new-asset") || strings.Contains(err.Error()+stderr, "secret-test") || out != "" || unexpectedRedirect.Load() {
		t.Fatalf("stdout = %q, stderr = %q, err = %v, redirected = %t", out, stderr, err, unexpectedRedirect.Load())
	}
}
