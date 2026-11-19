package printer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Table is an explicit view of an API response. RowsKey may be a dot-separated
// path; a top-level array needs no row key. Summary reads from the root object.
type Table struct {
	RowsKey string
	Columns []string
	Fields  map[string]Field
	Summary []string
	// More is the path to an API-provided next-page flag or cursor.
	More string
	// HasMore handles numbered lists where the response only exposes a total.
	HasMore func(root map[string]any) bool
	// Result is the confirmation for a successful mutation, including an empty body.
	Result string
}

type Field struct {
	Label  string
	Format func(value any, row map[string]any) string
	// MaxWidth limits prose, in runes. Zero uses the default for names and
	// descriptions; a negative value leaves the full text intact.
	MaxWidth int
}

func Print(cmd *cobra.Command, data []byte, table Table) error {
	data = bytes.TrimSpace(data)
	out := cmd.OutOrStdout()
	format, _ := cmd.Flags().GetString("output")
	if format == "json" {
		if len(data) == 0 {
			return nil
		}
		var formatted bytes.Buffer
		if err := json.Indent(&formatted, data, "", "  "); err != nil {
			return errors.New("API returned an invalid JSON response")
		}
		formatted.WriteByte('\n')
		_, err := formatted.WriteTo(out)
		return err
	}
	return printTable(out, data, table, colorEnabled(cmd, out))
}

func printTable(out io.Writer, data []byte, table Table, color bool) error {
	var value any
	if len(data) != 0 {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		if !json.Valid(data) || decoder.Decode(&value) != nil {
			return errors.New("API returned an invalid JSON response")
		}
	}
	if len(table.Columns) == 0 && table.Result == "" && len(table.Summary) == 0 {
		return errors.New("no table view is available; use -o json")
	}
	root, _ := value.(map[string]any)
	failed := table.Result != "" && failure(value, root)
	if table.RowsKey != "" {
		var rows any
		if value != nil {
			var exists bool
			rows, exists = lookup(value, table.RowsKey)
			if !exists && !failed {
				return fmt.Errorf("API response did not include %q rows; use -o json to inspect it", table.RowsKey)
			}
		}
		switch array := rows.(type) {
		case nil:
			value = []any(nil)
		case []any:
			value = array
		default:
			return fmt.Errorf("API response field %q is not an array", table.RowsKey)
		}
	}
	var output bytes.Buffer
	if table.Result != "" {
		if failed {
			fmt.Fprintln(&output, "API reported failure. Use -o json to inspect the response.")
		} else {
			fmt.Fprintln(&output, table.Result)
		}
	}
	if len(table.Columns) > 0 {
		if err := table.writeBody(&output, value, color); err != nil {
			return err
		}
	}
	if len(table.Summary) > 0 {
		writer := tabwriter.NewWriter(&output, 0, 4, 2, ' ', 0)
		table.writeFields(writer, root, table.Summary, true)
		if err := writer.Flush(); err != nil {
			return err
		}
	}
	var next any
	if table.More != "" {
		next, _ = lookup(root, table.More)
	}
	if more(next) || (table.HasMore != nil && table.HasMore(root)) {
		fmt.Fprintln(&output, "More results are available. Use -o json for pagination details.")
	}
	_, err := output.WriteTo(out)
	return err
}

func (t Table) writeBody(out *bytes.Buffer, value any, color bool) error {
	var body bytes.Buffer
	writer := tabwriter.NewWriter(&body, 0, 4, 2, ' ', 0)
	hasHeader := false
	switch v := value.(type) {
	case []any:
		if len(v) == 0 {
			if t.Result == "" {
				fmt.Fprintln(writer, "No results found.")
			}
			break
		}
		hasHeader = true
		labels := make([]string, len(t.Columns))
		for i, path := range t.Columns {
			labels[i] = t.label(path)
		}
		fmt.Fprintln(writer, strings.Join(labels, "\t"))
		values := make([]string, len(t.Columns))
		for _, item := range v {
			row, ok := item.(map[string]any)
			if !ok {
				return errors.New("API response included an invalid table row; use -o json to inspect it")
			}
			for i, path := range t.Columns {
				values[i] = display(t.cell(row, path))
			}
			fmt.Fprintln(writer, strings.Join(values, "\t"))
		}
	case map[string]any:
		t.writeFields(writer, v, t.Columns, t.Result != "")
	case nil:
		if t.Result == "" {
			fmt.Fprintln(writer, "No results found.")
		}
	default:
		if t.Result == "" {
			return errors.New("API response has no table fields; use -o json to inspect it")
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if color && hasHeader {
		// Style the header after layout so ANSI codes don't affect column widths.
		header, rows, _ := bytes.Cut(body.Bytes(), []byte("\n"))
		fmt.Fprintf(out, "\x1b[1m%s\x1b[0m\n%s", header, rows)
	} else {
		out.Write(body.Bytes())
	}
	return nil
}

func (t Table) writeFields(w io.Writer, row map[string]any, paths []string, omitMissing bool) {
	for _, path := range paths {
		text := t.cell(row, path)
		if text == "" && omitMissing {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\n", t.label(path), display(text))
	}
}

func lookup(value any, path string) (any, bool) {
	for part := range strings.SplitSeq(path, ".") {
		object, ok := value.(map[string]any)
		if !ok {
			return nil, false
		}
		value, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return value, true
}

func (t Table) label(path string) string {
	label := t.Fields[path].Label
	if label == "" {
		label = strings.ReplaceAll(leaf(path), "_", " ")
	}
	return strings.ToUpper(clean(label))
}

func leaf(path string) string {
	return path[strings.LastIndex(path, ".")+1:]
}

func (t Table) cell(row map[string]any, path string) string {
	value, _ := lookup(row, path)
	cfg := t.Fields[path]
	text := scalar(value)
	if cfg.Format != nil {
		text = cfg.Format(value, row)
	}
	text = clean(text)
	limit := cfg.MaxWidth
	if limit == 0 {
		switch leaf(path) {
		case "name", "title", "description":
			limit = 50
		}
	}
	if limit > 0 {
		runes := []rune(text)
		if len(runes) > limit {
			return string(runes[:limit-1]) + "…"
		}
	}
	return text
}

func display(text string) string {
	if text == "" {
		return "—"
	}
	return text
}

func scalar(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func more(value any) bool {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case bool:
		return v
	case json.Number:
		n, err := v.Int64()
		return err == nil && n > 0
	default:
		return false
	}
}

func failure(value any, root map[string]any) bool {
	if value == false {
		return true
	}
	for _, key := range []string{"success", "ok", "deleted", "uploaded", "status"} {
		if root[key] == false {
			return true
		}
	}
	switch strings.ToLower(strings.TrimSpace(scalar(root["status"]))) {
	case "error", "errored", "failed", "failure", "unsuccessful":
		return true
	}
	return hasError(root["error"]) || hasError(root["errors"])
}

func hasError(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.TrimSpace(v) != ""
	case map[string]any:
		return len(v) != 0
	case []any:
		return len(v) != 0
	}
	return false
}

func clean(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text)
}

func colorEnabled(cmd *cobra.Command, out io.Writer) bool {
	disabled, _ := cmd.Flags().GetBool("no-color")
	_, noColor := os.LookupEnv("NO_COLOR")
	file, ok := out.(*os.File)
	return !disabled && !noColor && ok && term.IsTerminal(int(file.Fd()))
}
