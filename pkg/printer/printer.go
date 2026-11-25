package printer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"gumlet/pkg/termui"

	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
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
}

func Print(cmd *cobra.Command, data []byte, table Table) error {
	data = bytes.TrimSpace(data)
	out := cmd.OutOrStdout()
	color := termui.ColorEnabled(cmd, out)
	format, _ := cmd.Flags().GetString("output")
	if format == "json" {
		return printJSON(out, data, color)
	}
	return printTable(out, data, table, color)
}

var jsonStyle = func() *pretty.Style {
	s := *pretty.TerminalStyle
	s.Key = [2]string{"\x1b[36m", "\x1b[0m"}
	s.True = [2]string{"\x1b[1;32m", "\x1b[0m"}
	s.False = [2]string{"\x1b[31m", "\x1b[0m"}
	s.Brackets = [2]string{"\x1b[2m", "\x1b[0m"}
	return &s
}()

func printJSON(out io.Writer, data []byte, color bool) error {
	if len(data) == 0 {
		return nil
	}
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, data, "", "  "); err != nil {
		return errors.New("API returned an invalid JSON response")
	}
	formatted.WriteByte('\n')
	if color {
		_, err := out.Write(pretty.Color(formatted.Bytes(), jsonStyle))
		return err
	}
	_, err := formatted.WriteTo(out)
	return err
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
	style := termui.NewStyle(color)
	var output bytes.Buffer
	if table.Result != "" {
		if failed {
			fmt.Fprintln(&output, style.Failure("API reported failure. Use -o json to inspect the response."))
		} else {
			fmt.Fprintln(&output, style.Success(table.Result))
		}
	}
	if len(table.Columns) > 0 {
		if err := table.writeBody(&output, value, style); err != nil {
			return err
		}
	}
	if len(table.Summary) > 0 {
		table.writeFields(&output, root, table.Summary, true, style)
	}
	var next any
	if table.More != "" {
		next, _ = lookup(root, table.More)
	}
	if more(next) || (table.HasMore != nil && table.HasMore(root)) {
		fmt.Fprintln(&output, style.Hint("More results are available. Use -o json for pagination details."))
	}
	_, err := output.WriteTo(out)
	return err
}

func (t Table) writeBody(out *bytes.Buffer, value any, style termui.Style) error {
	switch v := value.(type) {
	case []any:
		if len(v) == 0 {
			if t.Result == "" {
				fmt.Fprintln(out, style.Dim("No results found."))
			}
			return nil
		}
		headers := make([]string, len(t.Columns))
		for i, path := range t.Columns {
			headers[i] = style.BoldCyan(t.label(path))
		}

		rows := make([][]string, 0, len(v))
		for _, item := range v {
			row, ok := item.(map[string]any)
			if !ok {
				return errors.New("API response included an invalid table row; use -o json to inspect it")
			}
			cells := make([]string, len(t.Columns))
			for i, path := range t.Columns {
				text := display(t.cell(row, path))
				cells[i] = styleValue(style, path, text)
			}
			rows = append(rows, cells)
		}
		writeTable(out, headers, rows)
	case map[string]any:
		t.writeFields(out, v, t.Columns, t.Result != "", style)
	case nil:
		if t.Result == "" {
			fmt.Fprintln(out, style.Dim("No results found."))
		}
	default:
		if t.Result == "" {
			return errors.New("API response has no table fields; use -o json to inspect it")
		}
	}
	return nil
}

func (t Table) writeFields(out *bytes.Buffer, row map[string]any, paths []string, omitMissing bool, style termui.Style) {
	rows := make([][]string, 0, len(paths))
	for _, path := range paths {
		text := t.cell(row, path)
		if text == "" && omitMissing {
			continue
		}
		lbl := t.label(path)
		val := display(text)
		rows = append(rows, []string{
			style.BoldCyan(lbl),
			styleValue(style, path, val),
		})
	}
	writeTable(out, nil, rows)
}

func writeTable(out *bytes.Buffer, headers []string, rows [][]string) {
	if len(headers) == 0 && len(rows) == 0 {
		return
	}
	// lipgloss/table v1.1.0 miscomputes MaxHeight by -1 when Headers is empty
	// and borders are disabled, which drops the final row. Promote the first row
	// to Headers so all rows are rendered identically.
	if len(headers) == 0 {
		headers = rows[0]
		rows = rows[1:]
	}
	cols := len(headers)
	cellStyle := lipgloss.NewStyle().PaddingRight(2)
	lastColStyle := lipgloss.NewStyle()
	tbl := ltable.New().
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderHeader(false).
		BorderColumn(false).
		BorderRow(false).
		StyleFunc(func(_, col int) lipgloss.Style {
			if col < cols-1 {
				return cellStyle
			}
			return lastColStyle
		}).
		Headers(headers...).
		Rows(rows...)
	for line := range strings.SplitSeq(tbl.Render(), "\n") {
		out.WriteString(strings.TrimRight(line, " "))
		out.WriteByte('\n')
	}
}

func styleValue(style termui.Style, path, text string) string {
	if !style.Color() {
		return text
	}
	if text == "—" {
		return style.Dim(text)
	}
	if strings.HasPrefix(text, "https://") || strings.HasPrefix(text, "http://") {
		return style.Link(text)
	}
	key := leaf(path)
	lower := strings.ToLower(strings.TrimSpace(text))
	if key == "status" || key == "status_code" {
		switch lower {
		case "ready", "active", "paid", "complete", "completed", "published", "ok", "success":
			return style.Green(text)
		case "processing", "queued", "pre-queued", "preparing", "invited", "pending", "open", "created", "in_progress":
			return style.Yellow(text)
		case "error", "errored", "failed", "failure", "cancelled", "canceled", "deleted", "uncollectible":
			return style.Red(text)
		case "void", "inactive", "draft":
			return style.Dim(text)
		}
		if code, err := strconv.Atoi(lower); err == nil {
			switch {
			case code >= 200 && code < 300:
				return style.Green(text)
			case code >= 400:
				return style.Red(text)
			case code >= 300:
				return style.Yellow(text)
			}
		}
	}
	if lower == "true" || lower == "false" {
		if key == "unpaid" || key == "unpaid_invoices" {
			if lower == "true" {
				return style.Yellow(text)
			}
			return style.Green(text)
		}
		if lower == "true" {
			return style.Green(text)
		}
		return style.Dim(text)
	}
	return text
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
	switch leaf(path) {
	case "name", "title", "description":
		const maxProseWidth = 50
		if runes := []rune(text); len(runes) > maxProseWidth {
			return string(runes[:maxProseWidth-1]) + "…"
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
		return strconv.FormatBool(v)
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
