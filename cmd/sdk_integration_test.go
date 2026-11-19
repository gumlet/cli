package cmd_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gumlet/cmd"
	_ "gumlet/cmd/image/source"
	_ "gumlet/cmd/video/asset"
	_ "gumlet/cmd/video/playlist"
	_ "gumlet/cmd/video/workspace"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// Exercise the real command tree in a fresh process so Cobra, Viper, and the
// output streams don't leak state between command invocations.
func TestCLIProcess(t *testing.T) {
	if os.Getenv("GUMLET_CLI_TEST_PROCESS") != "1" {
		return
	}
	target, err := url.Parse(os.Getenv("GUMLET_CLI_TEST_API"))
	if err != nil {
		t.Fatal(err)
	}
	transport := &http.Transport{}
	http.DefaultClient.Transport = testTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme == "https" && r.URL.Host == "api.gumlet.com" {
			r = r.Clone(r.Context())
			r.URL.Scheme, r.URL.Host = target.Scheme, target.Host
			r.Host = target.Host
		}
		return transport.RoundTrip(r)
	})
	for i, arg := range os.Args {
		if arg == "--" {
			cmd.RootCmd.SetArgs(os.Args[i+1:])
			cmd.Execute()
			os.Exit(0)
		}
	}
	t.Fatal("missing command arguments")
}

func runCLI(t *testing.T, baseURL string, args ...string) ([]byte, error) {
	t.Helper()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gumlet.yaml"), []byte("api-key: test-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	process := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestCLIProcess$", "--"}, args...)...)
	process.Env = append(os.Environ(),
		"GUMLET_CLI_TEST_PROCESS=1", "GUMLET_CLI_TEST_API="+baseURL,
		"GUMLET_BASE_URL="+baseURL+"/unexpected/",
		"HOME="+home, "USERPROFILE="+home, "GUMLET_API-KEY=",
		"GUMLET_CUSTOM_HEADERS=X-Unexpected: injected", "API_KEY=wrong-key")
	return process.CombinedOutput()
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
		{"asset list", "GET", "/v1/video/assets/list/ws", "status=ready", "", `{"assets":[{"asset_id":"one","input":{"title":"Example"},"extra":1}],"total_asset_count":1}`, `{"assets":[{"asset_id":"one","title":"Example"}],"total_asset_count":1}`, []string{"video", "asset", "list", "--workspace-id", "ws", "--status", "ready", "--output", "json"}, 200},
		{"workspace create", "POST", "/v1/video/workspaces", "", `{"name":"Example"}`, `{"id":"ws"}`, `{"id":"ws"}`, []string{"video", "workspace", "create", "--name", "Example"}, 200},
		{"playlist create", "POST", "/v1/video/playlist", "", `{"collection_id":"ws","title":"Example","description":""}`, `{"id":"pl"}`, `{"id":"pl"}`, []string{"video", "playlist", "create", "--workspace-id", "ws", "--title", "Example"}, 200},
		{"playlist update", "POST", "/v1/video/playlist/pl", "", `{"channel_visibility":"public"}`, `{"id":"pl"}`, `{"id":"pl"}`, []string{"video", "playlist", "update", "--playlist-id", "pl", "--channel-visibility", "public"}, 200},
		{"playlist delete", "DELETE", "/v1/video/playlist/pl", "", "", `{"deleted":true}`, `{"deleted":true}`, []string{"video", "playlist", "delete", "--playlist-id", "pl"}, 200},
		{"empty delete", "DELETE", "/v1/video/assets/asset", "", "", "", "", []string{"video", "asset", "delete", "--asset-id", "asset"}, 204},
		{"legacy source config", "POST", "/v1/image/sources", "", `{"namespace":"test","type":"amazon","amazon":{"bucket_name":"example"}}`, `{"id":"src"}`, `{"id":"src"}`, []string{"image", "source", "add", "--namespace", "test", "--type", "amazon", "--config", `{"bucket_name":"example"}`}, 200},
		{"legacy source update", "POST", "/v1/image/sources/src", "", `{"type":"amazon","amazon":{"bucket_name":"example"}}`, `{"id":"src"}`, `{"id":"src"}`, []string{"image", "source", "update", "--source-id", "src", "--type", "amazon", "--config", `{"bucket_name":"example"}`}, 200},
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

			out, err := runCLI(t, server.URL, tc.args...)
			if tc.status >= 400 {
				if err == nil || !strings.Contains(string(out), "503") {
					t.Fatalf("expected API failure, got %v: %s", err, out)
				}
			} else if err != nil {
				t.Fatalf("command failed: %v: %s", err, out)
			} else if tc.output != "" {
				checkJSON(t, out, tc.output)
			}
			if got := requests.Load(); got != 1 {
				t.Errorf("request count = %d; want 1", got)
			}
		})
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
		checkJSON(t, body, `{"collection_id":"ws","format":"ABR","title":"sample.mp4","tag":["one","two"]}`)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"asset_id": "asset", "upload_url": storage.URL + "/file?signature=test"})
	}))
	defer api.Close()

	out, err := runCLI(t, api.URL, "video", "asset", "upload", "--file", file, "--workspace-id", "ws", "--tag", "one,two")
	if err != nil || !strings.Contains(string(out), "Upload complete! Asset ID: asset") || uploads.Load() != 1 {
		t.Fatalf("upload count=%d, err=%v, output=%s", uploads.Load(), err, out)
	}
}
