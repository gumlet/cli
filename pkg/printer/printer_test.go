package printer

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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
	table := Table{Columns: []string{"id", "name", "type"}}
	data := []byte(`[{"id":"abcdef0123456789","name":"Direct Upload","type":"direct-upload"}]`)
	if err := printTable(&out, data, table, true); err != nil {
		t.Fatal(err)
	}
	header, row, _ := strings.Cut(out.String(), "\n")
	if !strings.Contains(header, "\x1b[1m") {
		t.Fatalf("expected styled header: %q", header)
	}
	header = strings.NewReplacer("\x1b[1m", "", "\x1b[0m", "").Replace(header)
	for _, column := range [][2]string{{"NAME", "Direct Upload"}, {"TYPE", "direct-upload"}} {
		if strings.Index(header, column[0]) != strings.Index(row, column[1]) {
			t.Errorf("%s is misaligned:\n%s\n%s", column[0], header, row)
		}
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }
