package printer

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func TestPrintJSON(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.Flags().String("output", "json", "")
	if err := Print(cmd, []byte(`{"nested":{"count":9007199254740993}}`), Table{Columns: []string{"count"}}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "{\n  \"nested\": {\n    \"count\": 9007199254740993\n  }\n}\n" {
		t.Errorf("JSON output: %s", out.String())
	}

	out.Reset()
	if err := Print(cmd, []byte(`{"invalid":`), Table{}); err == nil || out.Len() != 0 {
		t.Errorf("invalid JSON: output=%q, err=%v", out.String(), err)
	}
}

func TestPrintTable(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.Flags().String("output", "table", "")
	table := Table{RowsKey: "rows", Columns: []string{"id", "input.title"}}
	if err := Print(cmd, []byte(`{"rows":[{"id":"asset","input":{"title":"hello\tworld"}}]}`), table); err != nil {
		t.Fatal(err)
	}
	if text := out.String(); !strings.Contains(text, "TITLE") || !strings.Contains(text, "hello world") {
		t.Errorf("table output: %s", text)
	}

	errSink := errors.New("write failed")
	cmd.SetOut(errorWriter{errSink})
	if err := Print(cmd, []byte(`{"rows":[{"id":"asset"}]}`), table); !errors.Is(err, errSink) {
		t.Errorf("writer error: %v", err)
	}
}

func TestStyledTableHeaderAlignment(t *testing.T) {
	var out bytes.Buffer
	table := Table{Columns: []string{"id", "name", "status", "type"}}
	data := []byte(`[{"id":"abcdef0123456789","name":"Direct Upload","status":"ready","type":"direct-upload"}]`)
	if err := printTable(&out, data, table, true); err != nil {
		t.Fatal(err)
	}
	header, row, _ := strings.Cut(out.String(), "\n")
	if !strings.Contains(header, "\x1b[1;36m") {
		t.Fatalf("expected styled header: %q", header)
	}
	if !strings.Contains(row, "\x1b[32mready\x1b[0m") {
		t.Fatalf("expected green status cell: %q", row)
	}
	plainHeader := stripANSI(header)
	plainRow := stripANSI(row)
	for _, column := range [][2]string{{"NAME", "Direct Upload"}, {"STATUS", "ready"}, {"TYPE", "direct-upload"}} {
		if strings.Index(plainHeader, column[0]) != strings.Index(plainRow, column[1]) {
			t.Errorf("%s is misaligned:\n%s\n%s", column[0], plainHeader, plainRow)
		}
	}
}

func TestKeyValueTableRendersAllRows(t *testing.T) {
	var out bytes.Buffer
	table := Table{Columns: []string{"id", "name", "status"}}
	data := []byte(`{"id":"asset-1","name":"日本語タイトル","status":"ready"}`)
	if err := printTable(&out, data, table, true); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3:\n%s", len(lines), out.String())
	}
	for i, want := range []string{"ID", "NAME", "STATUS"} {
		if !strings.HasPrefix(stripANSI(lines[i]), want) {
			t.Errorf("line %d = %q, want prefix %q", i, stripANSI(lines[i]), want)
		}
	}
}

func TestStyledJSONOutput(t *testing.T) {
	var out bytes.Buffer
	if err := printJSON(&out, []byte(`{"id":"asset-1","ready":true,"deleted":false}`), true); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, wantCode := range []string{"\x1b[36m\"id\"\x1b[0m", "\x1b[1;32mtrue\x1b[0m", "\x1b[31mfalse\x1b[0m"} {
		if !strings.Contains(got, wantCode) {
			t.Fatalf("expected %q in colorized JSON: %q", wantCode, got)
		}
	}
	want := "{\n  \"id\": \"asset-1\",\n  \"ready\": true,\n  \"deleted\": false\n}\n"
	if stripped := stripANSI(got); stripped != want {
		t.Errorf("stripped JSON = %q, want %q", stripped, want)
	}
}

func TestTableAllowlistAndPagination(t *testing.T) {
	const id = "an-important-identifier-that-is-longer-than-the-old-fifty-character-limit"
	const title = "This description is longer than the human-friendly width for prose fields."
	table := Table{
		RowsKey: "data.rows",
		Columns: []string{"id", "title", "nested.value", "roles", "count"},
		Fields: map[string]Field{
			"nested.value": {Label: "Selected value"},
			"roles":        {Format: JoinStrings(", ")},
		},
		More:    "data.next_page_token",
		Summary: []string{"unpaid"},
	}
	data := []byte(`{"data":{"rows":[{"id":"` + id + `","title":"` + title + `","nested":{"value":{"secret":"private-object"}},"roles":["editor","viewer"],"count":0}],"next_page_token":"private-cursor"},"unpaid":false}`)
	var out bytes.Buffer
	if err := printTable(&out, data, table, false); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{id, "SELECTED VALUE", "editor, viewer", "0\n", "—", "UNPAID", "false", "More results are available."} {
		if !strings.Contains(text, want) {
			t.Errorf("wanted %q in %q", want, text)
		}
	}
	if !strings.Contains(text, "…") || strings.Contains(text, title) || strings.Contains(text, "private-") {
		t.Errorf("unexpected contents: %q", text)
	}
}

func TestTableResult(t *testing.T) {
	const success = "Asset deleted."
	for _, tc := range []struct {
		name, body, want string
	}{
		{"empty", "", success},
		{"null", "null", success},
		{"success", `{"success":true}`, success},
		{"false", "false", "API reported failure."},
		{"false flag", `{"deleted":false}`, "API reported failure."},
		{"status", `{"status":"FAILED"}`, "API reported failure."},
		{"error object", `{"error":{"reason":"private-error-detail"}}`, "API reported failure."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := printTable(&out, []byte(tc.body), Table{Result: success}, false); err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(out.String(), tc.want) || strings.Contains(out.String(), "private-") {
				t.Errorf("got %q, want a safe %q", out.String(), tc.want)
			}
		})
	}
}

func TestTableRejectsUnconfiguredViewsAndMalformedRows(t *testing.T) {
	for _, tc := range []struct {
		data  string
		table Table
	}{
		{`{"sensitive":"private"}`, Table{}},
		{`{"rows":true}`, Table{RowsKey: "rows", Columns: []string{"id"}}},
		{`{"rows":[{"id":"one"},"private"]}`, Table{RowsKey: "rows", Columns: []string{"id"}}},
		{`{"unknown":[]}`, Table{RowsKey: "rows", Columns: []string{"id"}}},
	} {
		var out bytes.Buffer
		if err := printTable(&out, []byte(tc.data), tc.table, false); err == nil || out.Len() != 0 {
			t.Errorf("unsafe output %q, err=%v", out.String(), err)
		}
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }
