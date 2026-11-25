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
	const version = "1.2.3"
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
		if got := r.Header.Get("User-Agent"); got != "gumlet-cli/"+version {
			t.Errorf("unexpected API User-Agent: %q", got)
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

	root := cmd.NewRootCommand(cmd.Options{Version: version, NewClient: func() (*client.Client, error) {
		key, err := store.APIKey()
		if err != nil {
			return nil, err
		}
		return client.New(key, version, option.WithHTTPClient(httpClient)), nil
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
	const history = `{"stripe_invoices":[
		{"id":"invoice-usd","created":1704067200,"status":"paid","amount_due":9321648,"currency":"usd","invoice_pdf":"https://example.com/private-pdf"},
		{"id":"invoice-jpy","created":1704153600,"status":"open","amount_due":1000,"currency":"jpy"},
		{"id":"invoice-unspecified","created":1704240000,"status":"void","amount_due":0}
	],"unpaid_invoices":true}`
	const upcoming = `{"invoice_start_date":"01 Nov 2026","invoice_last_date":"30 Nov 2026","next_invoice_date":"01 Dec",
		"current_month_usage":"USD 120.00","upcoming_invoice":{"subtotal":"USD 100.00","lines":{"data":[{"description":"private-line-item"}]}}}`
	const user = `{"id":"u","name":"Ada","email":"viewer@example.com","roles":["admin","editor"],"timezone":"UTC","last_login_ip":"192.0.2.25","metadata":{"theme":"private-theme"}}`
	const viewers = `{"subscriptions":[{"name":"Viewer A","email":"viewer@example.com","status":"invited","invitation_link":"https://example.com/private-invite"}],"total_count":1}`
	const asset = `{"asset_id":"asset","status":"ready","input":{"title":"Film","secret":"private-input"},"output":{"playback_url":"https://example.com/assets/complete-long-playback-url-that-should-remain-copyable.m3u8","thumbnail_url":["https://example.com/thumbnail?token=private-thumbnail"]}}`
	const source = `{"id":"src","namespace":"test","type":"amazon","temp_cname":"test.gumlet.io","secure_token":"private-source"}`
	for _, tc := range []struct {
		name, method, path, query, body, response, output string
		args                                              []string
		status                                            int
	}{
		{"asset list", "GET", "/v1/video/assets/list/ws", "folder_id=&orderBy=asc&page_size=10&playlist_id=pl&sortBy=created_at&status=ready", "", `{"all_assets":[{"asset_id":"one","input":{"title":"Example"},"extra":1}],"total_asset_count":1}`, `{"all_assets":[{"asset_id":"one","input":{"title":"Example"},"extra":1}],"total_asset_count":1}`, []string{"video", "asset", "list", "--workspace-id", "ws", "--status", "ready", "--sort-by", "created_at", "--order-by", "asc", "--page-size", "10", "--folder-id=", "--playlist-id", "pl"}, 200},
		{"asset details", "GET", "/v1/video/assets/asset", "", "", asset, asset, []string{"video", "asset", "get", "--asset-id", "asset"}, 200},
		{"workspace create", "POST", "/v1/video/workspaces", "", `{"name":"Example"}`, `{"id":"ws"}`, `{"id":"ws"}`, []string{"video", "workspace", "create", "--name", "Example"}, 200},
		{"workspace list", "GET", "/v1/video/workspaces", "", "", `{"all_sources":[{"id":"ws","name":"Example","type":"direct-upload"}]}`, `{"all_sources":[{"id":"ws","name":"Example","type":"direct-upload"}]}`, []string{"video", "workspace", "list"}, 200},
		{"workspace details", "GET", "/v1/video/workspaces/ws", "", "", `{"id":"ws","name":"Example"}`, `{"id":"ws","name":"Example"}`, []string{"video", "workspace", "get", "--workspace-id", "ws"}, 200},
		{"workspace update", "POST", "/v1/video/workspaces/ws", "", `{"name":"New"}`, `{"id":"ws","name":"New"}`, `{"id":"ws","name":"New"}`, []string{"video", "workspace", "update", "--workspace-id", "ws", "--name", "New"}, 200},
		{"workspace delete", "DELETE", "/v1/video/workspaces/ws", "", "", "", "", []string{"video", "workspace", "delete", "--workspace-id", "ws"}, 204},
		{"playlist create", "POST", "/v1/video/playlist", "", `{"workspace_id":"ws","title":"Example"}`, `{"id":"pl"}`, `{"id":"pl"}`, []string{"video", "playlist", "create", "--workspace-id", "ws", "--title", "Example"}, 200},
		{"playlist list", "GET", "/v1/video/playlist", "workspace_id=ws", "", `[{"id":"pl"}]`, `[{"id":"pl"}]`, []string{"video", "playlist", "list", "--workspace-id", "ws"}, 200},
		{"playlist update", "POST", "/v1/video/playlist/pl", "", `{"channel_visibility":false}`, `{"id":"pl"}`, `{"id":"pl"}`, []string{"video", "playlist", "update", "--playlist-id", "pl", "--channel-visibility=false"}, 200},
		{"playlist delete", "DELETE", "/v1/video/playlist/pl", "", "", `{"deleted":true}`, `{"deleted":true}`, []string{"video", "playlist", "delete", "--playlist-id", "pl"}, 200},
		{"playlist assets", "GET", "/v1/video/playlist/pl/assets", "", "", `{"asset_list":[{"id":"asset","title":"Film","status":"ready","duration":60}],"has_next_page":true,"next_page":2}`, `{"asset_list":[{"id":"asset","title":"Film","status":"ready","duration":60}],"has_next_page":true,"next_page":2}`, []string{"video", "playlist", "get-assets", "--playlist-id", "pl"}, 200},
		{"playlist add asset", "POST", "/v1/video/playlist/pl/asset", "", `{"asset_list":[{"asset_id":"asset"}]}`, `{"success":true}`, `{"success":true}`, []string{"video", "playlist", "add-asset", "--playlist-id", "pl", "--asset-ids", "asset"}, 200},
		{"playlist remove asset", "DELETE", "/v1/video/playlist/pl/asset", "", `{"delete_list":["asset"]}`, `{"success":true}`, `{"success":true}`, []string{"video", "playlist", "remove-asset", "--playlist-id", "pl", "--asset-ids", "asset"}, 200},
		{"empty delete", "DELETE", "/v1/video/assets/asset", "", "", "", "", []string{"video", "asset", "delete", "--asset-id", "asset"}, 204},
		{"image source list", "GET", "/v1/image/sources", "", "", `{"all_sources":[{"id":"src","namespace":"test","type":"amazon","cname":["images.example.com"]}],"next_page_token":"private-image-cursor"}`, `{"all_sources":[{"id":"src","namespace":"test","type":"amazon","cname":["images.example.com"]}],"next_page_token":"private-image-cursor"}`, []string{"image", "source", "list"}, 200},
		{"image source get", "GET", "/v1/image/sources/src", "", "", source, source, []string{"image", "source", "get", "--source-id", "src"}, 200},
		{"image source delete", "DELETE", "/v1/image/sources/src", "", "", "", "", []string{"image", "source", "delete", "--source-id", "src"}, 204},
		{"legacy source config", "POST", "/v1/image/sources", "", `{"namespace":"test","type":"amazon","amazon":{"bucket_name":"example"}}`, `{"id":"src"}`, `{"id":"src"}`, []string{"image", "source", "add", "--namespace", "test", "--type", "amazon", "--config", `{"bucket_name":"example"}`}, 200},
		{"legacy source update", "POST", "/v1/image/sources/src", "", `{"type":"amazon","amazon":{"bucket_name":"example"}}`, `{"id":"src"}`, `{"id":"src"}`, []string{"image", "source", "update", "--source-id", "src", "--type", "amazon", "--config", `{"bucket_name":"example"}`}, 200},
		{"source purge", "POST", "/v1/image/purge/src", "", `{"paths":["/photo.jpg","banner.png"]}`, `{"status":"ok"}`, `{"status":"ok"}`, []string{"image", "purge", "--source-id", "src", "--urls", "/photo.jpg,banner.png"}, 200},
		{"current user", "GET", "/v1/user/data", "", "", user, user, []string{"user", "get"}, 200},
		{"org details", "GET", "/v1/org/data", "", "", `{"id":"org-1","org_email":"ops@example.com","plan":"business","plan_cycle":"monthly","unpaid":false,"sso_enabled":true,"created_at":1704067200000,"plan_data":{"video":{"display_name":"Video Pro"},"image":{"display_name":"Image Pro"}}}`, `{"id":"org-1","org_email":"ops@example.com","plan":"business","plan_cycle":"monthly","unpaid":false,"sso_enabled":true,"created_at":1704067200000,"plan_data":{"video":{"display_name":"Video Pro"},"image":{"display_name":"Image Pro"}}}`, []string{"org", "get"}, 200},
		{"audit logs", "POST", "/v1/user/audit-log", "", `{"date_range":{"start_at":"2026-08-01","end_at":"2026-08-31"},"page_number":2,"page_size":50,"user_email":["ops@example.com"],"activity_type":["video_uploaded","workspace_created"]}`, `{"pagination":{"page":2,"page_size":50,"total_pages":3,"total":120},"data":[{"timestamp":1704067200000,"activity":"Video Uploaded","action_user":"ops@example.com","status_code":200,"ip":"192.0.2.10","source":{"id":"asset-1","type":"asset"}}]}`, `{"pagination":{"page":2,"page_size":50,"total_pages":3,"total":120},"data":[{"timestamp":1704067200000,"activity":"Video Uploaded","action_user":"ops@example.com","status_code":200,"ip":"192.0.2.10","source":{"id":"asset-1","type":"asset"}}]}`, []string{"audit-log", "list", "--start-at", "2026-08-01", "--end-at", "2026-08-31", "--page-number", "2", "--page-size", "50", "--user-email", "ops@example.com", "--activity-type", "video_uploaded,workspace_created"}, 200},
		{"asset create from url", "POST", "/v1/video/assets", "", `{"workspace_id":"ws","input":"https://example.com/video.mp4","format":"ABR","title":"Demo","tag":["launch"],"folder":"fld-1","mp4_access":true}`, `{"asset_id":"asset-1","status":"pre-queued","workspace_id":"ws","input":{"title":"Demo"}}`, `{"asset_id":"asset-1","status":"pre-queued","workspace_id":"ws","input":{"title":"Demo"}}`, []string{"video", "asset", "create", "--workspace-id", "ws", "--input", "https://example.com/video.mp4", "--title", "Demo", "--tag", "launch", "--folder-id", "fld-1", "--mp4-access"}, 200},
		{"asset create with profile", "POST", "/v1/video/assets", "", `{"workspace_id":"ws","input":"https://example.com/video.mp4","profile_id":"prof-1"}`, `{"asset_id":"asset-2","status":"pre-queued"}`, `{"asset_id":"asset-2","status":"pre-queued"}`, []string{"video", "asset", "create", "--workspace-id", "ws", "--input", "https://example.com/video.mp4", "--profile-id", "prof-1"}, 200},
		{"asset thumbnail select", "POST", "/v1/video/assets/asset-1/thumbnail-select", "", `{"frame_at_second":2.5}`, `{"success":true,"asset_id":"asset-1","thumbnail_updated_at":1704067200000}`, `{"success":true,"asset_id":"asset-1","thumbnail_updated_at":1704067200000}`, []string{"video", "asset", "thumbnail-select", "--asset-id", "asset-1", "--frame-at-second", "2.5"}, 200},
		{"invoice history", "GET", "/v1/mixed/billing/invoice/history", "", "", history, history, []string{"billing", "invoice", "history"}, 200},
		{"empty invoice history", "GET", "/v1/mixed/billing/invoice/history", "", "", `{"stripe_invoices":[],"unpaid_invoices":false}`, `{"stripe_invoices":[],"unpaid_invoices":false}`, []string{"billing", "invoice", "history"}, 200},
		{"upcoming invoice", "GET", "/v1/mixed/billing/invoice/upcoming", "", "", upcoming, upcoming, []string{"billing", "invoice", "upcoming"}, 200},
		{"channel viewers page", "GET", "/v1/channel/ws/viewers", "page_number=2&page_size=25", "", viewers, viewers, []string{"video", "channel", "viewer", "list", "--workspace-id", "ws", "--page-number", "2", "--page-size", "25"}, 200},
		{"channel viewers cursor", "GET", "/v1/channel/ws/viewers", "next_page_token=viewer%2Bnext", "", `{"subscriptions":[],"next_page_token":"next"}`, `{"subscriptions":[],"next_page_token":"next"}`, []string{"video", "channel", "viewer", "list", "--workspace-id", "ws", "--next-page-token", "viewer+next"}, 200},
		{"channel viewer remove", "POST", "/v1/channel/ws/viewers/remove", "", `{"emails":["one@example.com","two@example.com"]}`, `{"success":true}`, `{"success":true}`, []string{"video", "channel", "viewer", "remove", "--workspace-id", "ws", "--emails", "one@example.com,two@example.com"}, 200},
		{"channel viewer remove rejected", "POST", "/v1/channel/ws/viewers/remove", "", `{"emails":["one@example.com"]}`, `{"success":false,"error":"private-removal-detail"}`, `{"success":false,"error":"private-removal-detail"}`, []string{"video", "channel", "viewer", "remove", "--workspace-id", "ws", "--emails", "one@example.com"}, 200},
		{"live workspace list", "GET", "/v1/video/sources/live", "", "", `{"all_live_sources":[{"id":"lw","name":"Studio"}]}`, `{"all_live_sources":[{"id":"lw","name":"Studio"}]}`, []string{"video", "live", "workspace", "list"}, 200},
		{"live workspace create", "POST", "/v1/video/sources/live", "", `{"name":"Studio"}`, `{"id":"lw"}`, `{"id":"lw"}`, []string{"video", "live", "workspace", "create", "--name", "Studio"}, 200},
		{"live workspace update", "POST", "/v1/video/sources/live/lw", "", `{"name":"New","video_source_id":"vod"}`, `{"id":"lw","video_workspace_id":"vod"}`, `{"id":"lw","video_workspace_id":"vod"}`, []string{"video", "live", "workspace", "update", "--live-workspace-id", "lw", "--name", "New", "--video-workspace-id", "vod"}, 200},
		{"live workspace delete", "DELETE", "/v1/video/sources/live/lw", "", "", `{"success":true}`, `{"success":true}`, []string{"video", "live", "workspace", "delete", "--live-workspace-id", "lw"}, 200},
		{"live asset create", "POST", "/v1/video/live/assets", "", `{"live_source_id":"lw","resolution":"720p","title":"Talk","mp4_access":false,"orientation":"landscape","start_at":"2026-12-01T15:00:00Z"}`, `{"live_asset_id":"la","stream_key":"fixture-key"}`, `{"live_asset_id":"la","stream_key":"fixture-key"}`, []string{"video", "live", "asset", "create", "--live-workspace-id", "lw", "--resolution", "720p", "--title", "Talk", "--mp4-access=false", "--orientation", "landscape", "--start-at", "2026-12-01T15:00:00Z"}, 200},
		{"live asset update", "POST", "/v1/video/live/assets/update", "", `{"live_asset_id":"la","title":"","start_at":"2026-12-01T15:00:00Z"}`, `{"live_asset_id":"la"}`, `{"live_asset_id":"la"}`, []string{"video", "live", "asset", "update", "--live-asset-id", "la", "--title=", "--start-at", "2026-12-01T15:00:00Z"}, 200},
		{"live asset list", "GET", "/v1/video/live/assets/list/lw", "next_page_token=next%2B1&page_size=10&status=created%2Cactive", "", `{"all_live_assets":[{"live_asset_id":"la"}],"next_page_token":"next"}`, `{"all_live_assets":[{"live_asset_id":"la"}],"next_page_token":"next"}`, []string{"video", "live", "asset", "list", "--live-workspace-id", "lw", "--status", "created,active", "--page-size", "10", "--next-page-token", "next+1"}, 200},
		{"live asset get", "GET", "/v1/video/live/assets/la", "", "", `{"status":"active"}`, `{"status":"active"}`, []string{"video", "live", "asset", "get", "--live-asset-id", "la"}, 200},
		{"live asset delete", "DELETE", "/v1/video/live/assets/la", "", "", "", "", []string{"video", "live", "asset", "delete", "--live-asset-id", "la"}, 204},
		{"live asset start", "POST", "/v1/video/live/assets/la/start", "", "", "", "", []string{"video", "live", "asset", "start", "--live-asset-id", "la"}, 200},
		{"live asset complete", "POST", "/v1/video/live/assets/la/complete", "", "", `{"status":"complete"}`, `{"status":"complete"}`, []string{"video", "live", "asset", "complete", "--live-asset-id", "la"}, 200},
		{"live asset history", "GET", "/v1/video/live/assets/la/history", "", "", `{"status":"complete","events":[{"status":"active"}]}`, `{"status":"complete","events":[{"status":"active"}]}`, []string{"video", "live", "asset", "history", "--live-asset-id", "la"}, 200},
		{"live thumbnail URLs", "POST", "/v1/video/live/assets/thumbnail/upload", "", `{"live_asset_id":"la","statuses":["preparing","end"]}`, `{"presigned_upload_urls":{"preparing":"https://example.com/upload?signature=fixture"}}`, `{"presigned_upload_urls":{"preparing":"https://example.com/upload?signature=fixture"}}`, []string{"video", "live", "asset", "thumbnail-upload-urls", "--live-asset-id", "la", "--statuses", "preparing,end"}, 200},
		{"failed mutation", "POST", "/v1/video/workspaces", "", `{"name":"Example"}`, `{"message":"unavailable"}`, "", []string{"video", "workspace", "create", "--name", "Example"}, 503},
		{"folder create", "POST", "/v1/video/workspaces/ws/folders", "", `{"name":"Clips"}`, `{"id":"folder"}`, `{"id":"folder"}`, []string{"video", "workspace", "folder", "create", "--workspace-id", "ws", "--name", "Clips"}, 200},
		{"nested folder create", "POST", "/v1/video/workspaces/ws/folders", "", `{"name":"Clips","parent_id":"parent"}`, `{"id":"folder"}`, `{"id":"folder"}`, []string{"video", "workspace", "folder", "create", "--workspace-id", "ws", "--name", "Clips", "--parent-id", "parent"}, 200},
		{"folder list", "GET", "/v1/video/workspaces/ws/folders", "parent_id=parent%2B1", "", `[{"id":"folder"}]`, `[{"id":"folder"}]`, []string{"video", "workspace", "folder", "list", "--workspace-id", "ws", "--parent-id", "parent+1"}, 200},
		{"root folder list", "GET", "/v1/video/workspaces/ws/folders", "parent_id=null", "", `[]`, `[]`, []string{"video", "workspace", "folder", "list", "--workspace-id", "ws", "--parent-id", "null"}, 200},
		{"folder get", "GET", "/v1/video/workspaces/ws/folders/folder", "", "", `{"id":"folder","path_names":["Library","Clips"],"asset_count":3}`, `{"id":"folder","path_names":["Library","Clips"],"asset_count":3}`, []string{"video", "workspace", "folder", "get", "--workspace-id", "ws", "--folder-id", "folder"}, 200},
		{"folder update", "POST", "/v1/video/workspaces/ws/folders/folder", "", `{"asset_ids":["one","two"]}`, `{"movedCount":2,"message":"success"}`, `{"movedCount":2,"message":"success"}`, []string{"video", "workspace", "folder", "update", "--workspace-id", "ws", "--folder-id", "folder", "--asset-ids", "one,two"}, 200},
		{"folder move to root", "POST", "/v1/video/workspaces/ws/folders/folder", "", `{"name":"Clips","parent_id":null}`, `{"id":"folder"}`, `{"id":"folder"}`, []string{"video", "workspace", "folder", "update", "--workspace-id", "ws", "--folder-id", "folder", "--name", "Clips", "--root"}, 200},
		{"folder delete", "DELETE", "/v1/video/workspaces/ws/folders/folder", "", "", "", "", []string{"video", "workspace", "folder", "delete", "--workspace-id", "ws", "--folder-id", "folder"}, 204},
		{"folder remove assets", "POST", "/v1/video/workspaces/ws/remove-assets-from-folder", "", `{"asset_ids":["one","two"]}`, "", "", []string{"video", "workspace", "folder", "remove-assets", "--workspace-id", "ws", "--asset-ids", "one,two"}, 204},
		{"asset update", "POST", "/v1/video/assets/update", "", `{"asset_id":"one","description":"","reprocess":false,"remove_subtitles":[]}`, `{"asset_id":"one","extra":true}`, `{"asset_id":"one","extra":true}`, []string{"video", "asset", "update", "--asset-id", "one", "--description=", "--reprocess=false", "--remove-subtitles="}, 200},
		{"recoverable list", "GET", "/v1/video/asset/recoverable/list", "page_number=2&page_size=25&workspace_id=ws", "", `{"all_assets":[],"total_asset_count":0}`, `{"all_assets":[],"total_asset_count":0}`, []string{"video", "asset", "list-recoverable", "--workspace-id", "ws", "--page-number", "2", "--page-size", "25"}, 200},
		{"recoverable next page", "GET", "/v1/video/asset/recoverable/list", "page_size=1&workspace_id=ws", "", `{"all_assets":[{"asset_id":"gone","title":"Removed","deleted_at":1704067200000}],"total_asset_count":2}`, `{"all_assets":[{"asset_id":"gone","title":"Removed","deleted_at":1704067200000}],"total_asset_count":2}`, []string{"video", "asset", "list-recoverable", "--workspace-id", "ws", "--page-size", "1"}, 200},
		{"asset recover", "POST", "/v1/video/asset/recover", "", `{"asset_id":"one"}`, "", "", []string{"video", "asset", "recover", "--asset-id", "one"}, 204},
		{"bulk tag add", "POST", "/v1/video/assets/bulk/tag", "", `{"workspace_id":"ws","asset_list":["one","two"],"add_tags":["promo"],"remove_tags":[]}`, `{"ok":true}`, `{"ok":true}`, []string{"video", "asset", "bulk-tag", "--workspace-id", "ws", "--asset-ids", "one,two", "--add-tags", "promo"}, 200},
		{"bulk tag remove", "POST", "/v1/video/assets/bulk/tag", "", `{"workspace_id":"ws","asset_list":["one"],"add_tags":[],"remove_tags":["promo"]}`, "", "", []string{"video", "asset", "bulk-tag", "--workspace-id", "ws", "--asset-ids", "one", "--remove-tags", "promo"}, 204},
		{"bulk delete", "DELETE", "/v1/video/assets/bulk/delete", "", `{"workspace_id":"ws","asset_list":["one","two"]}`, "", "", []string{"video", "asset", "bulk-delete", "--workspace-id", "ws", "--asset-ids", "one,two"}, 204},
		{"playlist reorder move", "POST", "/v1/video/playlists/pl/reorder", "", `{"asset_id":"one","asset_position":0,"page_number":2,"page_size":20}`, `{"ok":true}`, `{"ok":true}`, []string{"video", "playlist", "reorder", "--playlist-id", "pl", "--asset-id", "one", "--asset-position", "0", "--page-number", "2", "--page-size", "20"}, 200},
		{"playlist reorder sort", "POST", "/v1/video/playlists/pl/reorder", "", `{"sort_by":"created_at","sort_order":"desc"}`, `{"ok":true}`, `{"ok":true}`, []string{"video", "playlist", "reorder", "--playlist-id", "pl", "--sort-by", "created_at", "--sort-order", "desc"}, 200},
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
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if tc.body != "" {
					checkJSON(t, body, tc.body)
				} else if len(body) != 0 {
					t.Errorf("unexpected request body: %s", body)
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
			wantRequests := int32(1)
			if tc.status < 400 {
				wantRequests++
				args := append(append([]string{}, tc.args...), "-o", "table")
				table, diagnostics, err := runCLI(t, server.URL, "", args...)
				if err != nil || diagnostics != "" || strings.TrimSpace(table) == "" {
					t.Fatalf("table output=%q, diagnostics=%q, err=%v", table, diagnostics, err)
				}
				assertTableView(t, tc.name, table)
			}
			if got := requests.Load(); got != wantRequests {
				t.Errorf("request count = %d; want %d", got, wantRequests)
			}
		})
	}
}

func assertTableView(t *testing.T, fixture, table string) {
	t.Helper()
	want := map[string][]string{
		"asset details":                  {"Film", "ready", "https://example.com/assets/complete-long-playback-url-that-should-remain-copyable.m3u8"},
		"empty delete":                   {"Asset deleted."},
		"recoverable next page":          {"Removed", "2024-01-01 00:00 UTC", "More results are available."},
		"folder get":                     {"Library / Clips", "ASSET COUNT"},
		"folder update":                  {"Folder updated.", "ASSETS MOVED"},
		"invoice history":                {"DATE (UTC)", "AMOUNT DUE", "2024-01-01", "USD 93216.48", "JPY 1000", "0 minor units", "invoice-usd", "invoice-jpy", "invoice-unspecified", "UNPAID INVOICES", "true"},
		"empty invoice history":          {"No results found.", "UNPAID INVOICES", "false"},
		"upcoming invoice":               {"01 Nov 2026", "30 Nov 2026", "01 Dec", "USD 120.00", "USD 100.00"},
		"current user":                   {"Ada", "viewer@example.com", "admin, editor", "UTC"},
		"channel viewers page":           {"Viewer A", "viewer@example.com", "invited"},
		"channel viewers cursor":         {"No results found.", "More results are available."},
		"channel viewer remove rejected": {"API reported failure."},
		"playlist assets":                {"Film", "ready", "More results are available."},
		"image source list":              {"images.example.com", "More results are available."},
		"image source get":               {"test.gumlet.io"},
		"live asset list":                {"More results are available."},
		"live asset start":               {"Live stream start requested."},
	}
	for _, part := range want[fixture] {
		if !strings.Contains(table, part) {
			t.Errorf("table is missing %q:\n%s", part, table)
		}
	}
	for _, part := range []string{"private-", "192.0.2.25", "fixture-key", "signature=fixture", `[{"`, `{"`} {
		if strings.Contains(table, part) {
			t.Errorf("unexpected raw or sensitive value %q:\n%s", part, table)
		}
	}
	if fixture == "invoice history" {
		invoiceRows := 0
		for line := range strings.SplitSeq(table, "\n") {
			if strings.Contains(line, "invoice-") {
				invoiceRows++
			}
		}
		if invoiceRows != 3 {
			t.Errorf("wanted three invoice rows, got %d:\n%s", invoiceRows, table)
		}
	}
	if fixture == "channel viewer remove rejected" && strings.Contains(table, "removed") {
		t.Errorf("failure printed a success message:\n%s", table)
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

func TestAssetUpdateJSONInput(t *testing.T) {
	const metadata = `{"internal_id":9007199254740993}`
	const actions = `[{"text":"Learn more","start_time":0,"position_from_right":"12"}]`
	dir := t.TempDir()
	metaPath, actionsPath := filepath.Join(dir, "metadata.json"), filepath.Join(dir, "actions.json")
	for path, data := range map[string]string{metaPath: metadata, actionsPath: actions} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct{ metaPath, actionsPath, input, wantActions string }{
		{"-", actionsPath, metadata, actions},
		{metaPath, "-", "[]", "[]"},
	} {
		var requests atomic.Int32
		api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			var gotMetadata string
			if err := json.Unmarshal(body["metadata"], &gotMetadata); err != nil || gotMetadata != metadata {
				t.Errorf("metadata = %q, err = %v", gotMetadata, err)
			}
			checkJSON(t, body["call_to_actions"], tc.wantActions)
			w.WriteHeader(http.StatusNoContent)
		}))
		out, _, err := runCLI(t, api.URL, tc.input, "video", "asset", "update", "--asset-id", "one",
			"--metadata-file", tc.metaPath, "--call-to-actions-file", tc.actionsPath)
		api.Close()
		if err != nil || out != "" || requests.Load() != 1 {
			t.Fatalf("stdout = %q, err = %v, requests = %d", out, err, requests.Load())
		}
	}
}

func TestChannelViewerInviteInput(t *testing.T) {
	const users = `[{"email":"one@example.com","name":"One"},{"email":"two@example.com","name":"Two"}]`
	path := filepath.Join(t.TempDir(), "viewers.json")
	if err := os.WriteFile(path, []byte(users), 0600); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != "POST" || r.URL.Path != "/v1/channel/ws/viewers/invite" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		checkJSON(t, body, `{"users":`+users+`}`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer api.Close()
	for _, filename := range []string{path, "-"} {
		out, stderr, err := runCLI(t, api.URL, users, "video", "channel", "viewer", "invite",
			"--workspace-id", "ws", "--viewers-file", filename)
		if err != nil || stderr != "" {
			t.Fatalf("invite failed: %v (%s)", err, stderr)
		}
		checkJSON(t, []byte(out), `{"success":true}`)
	}
	out, stderr, err := runCLI(t, api.URL, users, "video", "channel", "viewer", "invite",
		"--workspace-id", "ws", "--viewers-file", "-", "-o", "table")
	if err != nil || stderr != "" || !strings.Contains(strings.ToLower(out), "invit") {
		t.Fatalf("invite table: stdout=%q, stderr=%q, err=%v", out, stderr, err)
	}
	if got := requests.Load(); got != 3 {
		t.Errorf("request count = %d; want 3", got)
	}
}

func TestLiveCredentialOutput(t *testing.T) {
	const secret = "fake-stream-key"
	const signed = "https://example.com/upload?signature=fake-signed-url"
	for _, tc := range []struct {
		name, response, want string
		args                 []string
		status               int
	}{
		{"status table", `{"live_asset_id":"la","status":"active","stream_key":"` + secret + `","stream_url":"rtmps://example.com/` + secret + `"}`, "active", []string{"get", "--live-asset-id", "la"}, 200},
		{"list table", `{"all_live_assets":[{"live_asset_id":"la","status":"active","stream_key":"` + secret + `","updated_at":{"unexpected":"` + secret + `"}}]}`, "active", []string{"list", "--live-workspace-id", "lw"}, 200},
		{"thumbnail table", `{"presigned_upload_urls":{"end":"` + signed + `"},"message":"` + signed + `"}`, "Use --output json", []string{"thumbnail-upload-urls", "--live-asset-id", "la", "--statuses", "end"}, 200},
		{"live failure", `{"message":"` + secret + ` ` + signed + `"}`, "HTTP 403", []string{"get", "--live-asset-id", "la"}, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.response)
			}))
			defer api.Close()
			args := append([]string{"video", "live", "asset", "--output", "table"}, tc.args...)
			out, stderr, err := runCLI(t, api.URL, "", args...)
			got := out + stderr + fmt.Sprint(err)
			if (err != nil) != (tc.status >= 400) || !strings.Contains(got, tc.want) ||
				strings.Contains(got, secret) || strings.Contains(got, "fake-signed-url") {
				t.Fatalf("unexpected live command output: %q", got)
			}
		})
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
		{[]string{"image", "purge", "--subdomain", "legacy"}, "unknown flag: --subdomain"},
		{[]string{"image", "purge", "--source-id", "src"}, "urls"},
		{[]string{"image", "purge", "--source-id", "src", "--urls="}, "at least one image path"},
		{[]string{"image", "purge", "--source-id", "src", "--urls", "/photo.jpg?width=100"}, "without a query"},
		{[]string{"image", "purge", "--source-id", "src", "--urls", "https://example.com/photo.jpg"}, "must be an image path"},
		{[]string{"video", "channel", "viewer", "list", "--workspace-id", "ws", "--page-size", "101"}, "between 1 and 100"},
		{[]string{"video", "channel", "viewer", "list", "--workspace-id", "ws", "--page-number", "2", "--next-page-token", "next"}, "page-number"},
		{[]string{"video", "channel", "viewer", "invite", "--workspace-id", "ws", "--viewers-file", "-"}, "JSON array"},
		{[]string{"video", "channel", "viewer", "remove", "--workspace-id", "ws", "--emails", "invalid"}, "bare email address"},
		{[]string{"video", "live", "workspace", "update", "--live-workspace-id", "lw"}, "at least one"},
		{[]string{"video", "live", "asset", "create", "--live-workspace-id", "lw", "--resolution", "4k"}, "--resolution"},
		{[]string{"video", "live", "asset", "create", "--live-workspace-id", "lw", "--resolution", "720p", "--orientation", "invalid"}, "--orientation"},
		{[]string{"video", "live", "asset", "update", "--live-asset-id", "la", "--start-at", "not-a-date"}, "RFC 3339"},
		{[]string{"video", "live", "asset", "list", "--live-workspace-id", "lw", "--status", "unknown"}, "--status accepts only"},
		{[]string{"video", "live", "asset", "list", "--live-workspace-id", "lw", "--page-number", "0"}, "at least 1"},
		{[]string{"video", "live", "asset", "thumbnail-upload-urls", "--live-asset-id", "la", "--statuses", "complete"}, "--statuses accepts only"},
		{[]string{"image", "source", "add", "--namespace", "test", "--type", "amazon", "--config", "null"}, "valid JSON object"},
		{[]string{"image", "source", "update", "--source-id", "src", "--config", "{}"}, "type"},
		{[]string{"video", "workspace", "folder", "list"}, "workspace-id"},
		{[]string{"video", "workspace", "folder", "list", "--workspace-id", "ws", "-o", "csv"}, "invalid --output"},
		{[]string{"video", "workspace", "folder", "update", "--workspace-id", "ws", "--folder-id", "f", "--root"}, "specify --name or --asset-ids"},
		{[]string{"video", "workspace", "folder", "update", "--workspace-id", "ws", "--folder-id", "f", "--asset-ids", "one", "--root"}, "--root requires --name"},
		{[]string{"video", "workspace", "folder", "update", "--workspace-id", "ws", "--folder-id", "f", "--name", "Clips", "--root", "--parent-id", "p"}, "cannot be combined"},
		{[]string{"video", "workspace", "folder", "remove-assets", "--workspace-id", "ws", "--asset-ids", "one,,two"}, "empty IDs"},
		{[]string{"video", "asset", "update", "--asset-id", "one"}, "at least one"},
		{[]string{"video", "asset", "update", "--asset-id", "one", "--metadata-file", "-", "--call-to-actions-file", "-"}, "cannot both read stdin"},
		{[]string{"video", "asset", "update", "--asset-id", "one", "--metadata-file", "-"}, "valid JSON object"},
		{[]string{"video", "asset", "list-recoverable", "--workspace-id", "ws", "--page-size", "101"}, "between 1 and 100"},
		{[]string{"video", "asset", "bulk-delete", "--workspace-id", "ws", "--asset-ids="}, "at least one asset ID"},
		{[]string{"video", "asset", "bulk-tag", "--workspace-id", "ws", "--asset-ids", "one"}, "at least one tag"},
		{[]string{"video", "playlist", "reorder", "--playlist-id", "pl", "--sort-by", "title"}, "sort-order"},
		{[]string{"video", "playlist", "reorder", "--playlist-id", "pl", "--sort-by", "title", "--sort-order", "asc", "--asset-id", "one"}, "asset-id"},
		{[]string{"video", "playlist", "reorder", "--playlist-id", "pl", "--asset-id", "one", "--page-number", "1", "--page-size", "20", "--asset-position", "20"}, "--asset-position"},
		{[]string{"video", "playlist", "reorder", "--playlist-id", "pl", "--sort-by", "status", "--sort-order", "asc"}, "invalid --sort-by"},
		{[]string{"video", "asset", "create", "--workspace-id", " ", "--input", "https://example.com/v.mp4"}, "--workspace-id"},
		{[]string{"video", "asset", "create", "--workspace-id", "ws", "--input", " "}, "--input"},
		{[]string{"video", "asset", "create", "--workspace-id", "ws", "--input", "https://example.com/v.mp4", "--format", "HLS"}, "invalid --format"},
		{[]string{"video", "asset", "thumbnail-upload", "--asset-id", " ", "--file", "thumb.png"}, "--asset-id"},
		{[]string{"video", "asset", "thumbnail-select", "--asset-id", " ", "--frame-at-second", "1"}, "--asset-id"},
		{[]string{"video", "asset", "thumbnail-select", "--asset-id", "asset-1", "--frame-at-second", "-1"}, "non-negative"},
		{[]string{"audit-log", "list", "--start-at", "2026/08/01", "--end-at", "2026-08-31"}, "YYYY-MM-DD"},
		{[]string{"audit-log", "list", "--start-at", "2026-08-31", "--end-at", "2026-08-01"}, "before --start-at"},
		{[]string{"audit-log", "list", "--start-at", "2026-08-01", "--end-at", "2026-08-31", "--page-size", "5"}, "between 10 and 500"},
		{[]string{"audit-log", "list", "--start-at", "2026-08-01", "--end-at", "2026-08-31", "--user-email", "not-an-email"}, "bare email"},
		{[]string{"audit-log", "list", "--start-at", "2026-08-01", "--end-at", "2026-08-31", "--activity-type", "unknown_event"}, "invalid --activity-type"},
	} {
		root := cmd.NewRootCommand(cmd.Options{NewClient: func() (*client.Client, error) {
			t.Fatal("created API client for invalid input")
			return nil, nil
		}})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetIn(strings.NewReader(""))
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
	table, stderr, err := runCLI(t, api.URL, "", "video", "asset", "upload", "--file", file, "--workspace-id", "ws", "--tag", "one,two", "-o", "table")
	if err != nil || uploads.Load() != 2 || !strings.Contains(stderr, "Uploading video") ||
		!strings.Contains(table, "asset") || !strings.Contains(strings.ToLower(table), "upload") ||
		strings.Contains(table, "signature=") {
		t.Fatalf("upload count=%d, err=%v, table=%q", uploads.Load(), err, table)
	}
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

func TestSDKThumbnailUpload(t *testing.T) {
	img := "fake png bytes"
	file := filepath.Join(t.TempDir(), "thumb.png")
	if err := os.WriteFile(file, []byte(img), 0600); err != nil {
		t.Fatal(err)
	}
	var uploads atomic.Int32
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploads.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != img || r.Method != "PUT" || r.URL.RawQuery != "signature=thumb" || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected thumbnail upload: %s %s, auth=%t, body=%q, err=%v", r.Method, r.URL, r.Header.Get("Authorization") != "", body, err)
		}
	}))
	defer storage.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/video/assets/asset-1/thumbnail" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected thumbnail session request: %s %s", r.Method, r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"asset_id":             "asset-1",
			"upload_url":           storage.URL + "/thumb.png?signature=thumb",
			"thumbnail_updated_at": 1704067200000,
		})
	}))
	defer api.Close()

	out, stderr, err := runCLI(t, api.URL, "", "video", "asset", "thumbnail-upload", "--asset-id", "asset-1", "--file", file)
	if err != nil || uploads.Load() != 1 || !strings.Contains(stderr, "Uploading thumbnail") {
		t.Fatalf("upload count=%d, err=%v, output=%s", uploads.Load(), err, out)
	}
	checkJSON(t, []byte(out), `{"asset_id":"asset-1","uploaded":true,"thumbnail_updated_at":1704067200000}`)
	if strings.Contains(out, "signature=") {
		t.Fatalf("signed URL leaked in output: %s", out)
	}

	table, _, err := runCLI(t, api.URL, "", "video", "asset", "thumbnail-upload", "--asset-id", "asset-1", "--file", file, "-o", "table")
	if err != nil || uploads.Load() != 2 || !strings.Contains(table, "Thumbnail uploaded.") || !strings.Contains(table, "asset-1") || strings.Contains(table, "signature=") {
		t.Fatalf("upload count=%d, err=%v, table=%q", uploads.Load(), err, table)
	}
}

func TestOrgAndAuditLogTableOutput(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/org/data":
			_, _ = io.WriteString(w, `{"id":"org-1","org_email":"ops@example.com","plan":"business","plan_cycle":"monthly","unpaid":false,"sso_enabled":true,"created_at":1704067200000,"plan_data":{"video":{"display_name":"Video Pro"},"image":{"display_name":"Image Pro"}}}`)
		case "/v1/user/audit-log":
			_, _ = io.WriteString(w, `{"pagination":{"page":1,"page_size":10,"total_pages":2,"total":15},"data":[{"timestamp":1704067200000,"activity":"Video Uploaded","action_user":"ops@example.com","status_code":200,"ip":"192.0.2.99","source":{"id":"asset-1","type":"asset"}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()

	orgOut, _, err := runCLI(t, api.URL, "", "org", "get", "-o", "table")
	if err != nil || !strings.Contains(orgOut, "org-1") || !strings.Contains(orgOut, "Video Pro") || !strings.Contains(orgOut, "2024-01-01 00:00 UTC") {
		t.Fatalf("org table = %q, err = %v", orgOut, err)
	}

	auditOut, _, err := runCLI(t, api.URL, "", "audit-log", "list", "--start-at", "2026-08-01", "--end-at", "2026-08-31", "-o", "table")
	if err != nil || !strings.Contains(auditOut, "Video Uploaded") || !strings.Contains(auditOut, "More results are available") || strings.Contains(auditOut, "192.0.2.99") {
		t.Fatalf("audit-log table = %q, err = %v", auditOut, err)
	}
}
