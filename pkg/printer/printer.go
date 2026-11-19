package printer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Table describes the fields to display and, for a list envelope, its row key.
// A top-level array needs no row key.
type Table struct {
	RowsKey string
	Columns []string
}

func Print(cmd *cobra.Command, data []byte, table Table) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil
	}
	out := cmd.OutOrStdout()
	format, _ := cmd.Flags().GetString("output")
	if format == "json" {
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
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if !json.Valid(data) || decoder.Decode(&value) != nil {
		return errors.New("API returned an invalid JSON response")
	}
	if table.RowsKey != "" {
		if envelope, ok := value.(map[string]any); ok {
			rows, exists := envelope[table.RowsKey]
			if !exists {
				return fmt.Errorf("API response did not include %q rows; use -o json to inspect it", table.RowsKey)
			}
			if rows == nil {
				value = []any{}
			} else if array, ok := rows.([]any); ok {
				value = array
			} else {
				return fmt.Errorf("API response field %q is not an array", table.RowsKey)
			}
		}
	}
	var formatted bytes.Buffer
	writer := tabwriter.NewWriter(&formatted, 0, 4, 2, ' ', 0)
	hasHeader := false
	switch v := value.(type) {
	case []any:
		if len(v) == 0 {
			_, err := fmt.Fprintln(out, "No results found.")
			return err
		}
		columns := table.Columns
		if len(columns) == 0 {
			if row, ok := v[0].(map[string]any); ok {
				columns = keys(row)
			}
		}
		if len(columns) > 0 {
			hasHeader = true
			labels := make([]string, len(columns))
			for i, name := range columns {
				labels[i] = heading(name)
			}
			_, _ = fmt.Fprintln(writer, strings.Join(labels, "\t"))
		}
		for _, row := range v {
			if len(columns) == 0 {
				_, _ = fmt.Fprintln(writer, cell(row))
				continue
			}
			values := make([]string, len(columns))
			for i, column := range columns {
				values[i] = cell(field(row, column))
			}
			_, _ = fmt.Fprintln(writer, strings.Join(values, "\t"))
		}
	case map[string]any:
		columns := table.Columns
		if len(columns) == 0 {
			columns = keys(v)
		}
		for _, column := range columns {
			_, _ = fmt.Fprintf(writer, "%s\t%s\n", heading(column), cell(field(v, column)))
		}
	default:
		_, _ = fmt.Fprintln(writer, cell(v))
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if color && hasHeader {
		// Style the header after layout so ANSI codes don't affect column widths.
		header, rows, _ := bytes.Cut(formatted.Bytes(), []byte("\n"))
		_, err := fmt.Fprintf(out, "\x1b[1m%s\x1b[0m\n%s", header, rows)
		return err
	}
	_, err := formatted.WriteTo(out)
	return err
}

func keys(value map[string]any) []string {
	out := make([]string, 0, len(value))
	for key := range value {
		out = append(out, key)
	}
	slices.Sort(out)
	return out
}

func field(value any, path string) any {
	for part := range strings.SplitSeq(path, ".") {
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = object[part]
	}
	return value
}

func heading(path string) string {
	path = path[strings.LastIndex(path, ".")+1:]
	return strings.ToUpper(clean(strings.ReplaceAll(path, "_", " ")))
}

func cell(value any) string {
	if value == nil || value == "" {
		return "—"
	}
	text, ok := value.(string)
	if !ok {
		encoded, _ := json.Marshal(value)
		text = string(encoded)
	}
	runes := []rune(clean(text))
	if len(runes) > 50 {
		return string(runes[:49]) + "…"
	}
	return string(runes)
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
