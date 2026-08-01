package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gumlet/pkg/ui"
)

const (
	maxCellWidth = 40
	colGap       = "  "
)

// Print outputs data in the specified format ("json" or "table").
// fields optionally limits which keys are shown; all keys shown if empty.
func Print(data []byte, format string, fields ...string) error {
	switch format {
	case "table":
		return printTable(data, fields)
	default:
		return printJSON(data, fields)
	}
}

func printJSON(data []byte, fields []string) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return nil
	}
	if len(fields) > 0 {
		v = filterByFields(v, fields)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(string(out))
	return nil
}

// filterByFields filters a decoded JSON value to only the requested field paths,
// flattening dot-notation paths to their label (last segment).
func filterByFields(v interface{}, fields []string) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(fields))
		for _, f := range fields {
			extracted := extractField(val, f)
			if extracted != nil {
				out[fieldLabel(f)] = extracted
			}
		}
		if len(out) == 0 {
			for k, child := range val {
				val[k] = filterByFields(child, fields)
			}
			return val
		}
		return out
	case []interface{}:
		for i, item := range val {
			val[i] = filterByFields(item, fields)
		}
		return val
	}
	return v
}

func printTable(data []byte, fields []string) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Println(string(data))
		return nil
	}

	switch v := raw.(type) {
	case []interface{}:
		return renderSlice(v, fields)
	case map[string]interface{}:
		if len(fields) > 0 {
			for _, f := range fields {
				if extractField(v, f) != nil {
					return renderKeyValue(v, fields)
				}
			}
		}
		var best []interface{}
		bestScore := -1
		for _, val := range v {
			arr, ok := val.([]interface{})
			if !ok || len(arr) == 0 {
				continue
			}
			score := 0
			if len(fields) > 0 {
				if firstObj, ok := arr[0].(map[string]interface{}); ok {
					for _, f := range fields {
						if extractField(firstObj, f) != nil {
							score++
						}
					}
				}
			}
			if score > bestScore || (score == bestScore && len(arr) > len(best)) {
				best = arr
				bestScore = score
			}
		}
		if best != nil {
			return renderSlice(best, fields)
		}
		return renderKeyValue(v, fields)
	default:
		fmt.Println(string(data))
	}
	return nil
}

func renderSlice(rows []interface{}, fields []string) error {
	if len(rows) == 0 {
		fmt.Println(ui.EmptyState())
		return nil
	}

	firstObj, ok := rows[0].(map[string]interface{})
	if !ok {
		for _, r := range rows {
			fmt.Println(r)
		}
		return nil
	}

	headers := resolveFields(firstObj, fields)
	headerLabels := make([]string, len(headers))
	for i, h := range headers {
		headerLabels[i] = humanHeader(fieldLabel(h))
	}

	plainRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		obj, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		cells := make([]string, len(headers))
		for i, h := range headers {
			cells[i] = plainCell(extractField(obj, h))
		}
		plainRows = append(plainRows, cells)
	}

	widths := columnWidths(headerLabels, plainRows)
	printTableRow(styleHeaders(headerLabels, widths))
	printTableRow(styleSeparators(widths))
	for _, cells := range plainRows {
		printTableRow(styleCells(headers, cells, widths))
	}

	fmt.Println()
	fmt.Println(ui.CountLabel(len(plainRows), "result", "results"))
	return nil
}

func renderKeyValue(obj map[string]interface{}, fields []string) error {
	keys := resolveFields(obj, fields)
	labels := make([]string, len(keys))
	values := make([]string, len(keys))
	for i, k := range keys {
		labels[i] = humanHeader(fieldLabel(k))
		values[i] = plainCell(extractField(obj, k))
	}

	labelWidth := maxLen(labels)
	for i, label := range labels {
		paddedLabel := ui.PadRight(label, labelWidth)
		styledLabel := ui.Cyan(os.Stdout, paddedLabel)
		plainValue := values[i]
		if plainValue == "" {
			plainValue = "—"
		}
		styledValue := ui.StyleCell(fieldLabel(keys[i]), plainValue)
		fmt.Println(styledLabel + colGap + styledValue)
	}
	return nil
}

func plainCell(v interface{}) string {
	raw := cellString(v)
	if raw == "" {
		return "—"
	}
	return truncate(raw, maxCellWidth)
}

func columnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	return widths
}

func maxLen(items []string) int {
	max := 0
	for _, s := range items {
		if len(s) > max {
			max = len(s)
		}
	}
	return max
}

func printTableRow(cells []string) {
	fmt.Println(strings.Join(cells, colGap))
}

// styleHeaders returns bold, width-padded header cells.
func styleHeaders(labels []string, widths []int) []string {
	out := make([]string, len(labels))
	for i, label := range labels {
		out[i] = ui.Bold(os.Stdout, ui.PadRight(label, widths[i]))
	}
	return out
}

func styleSeparators(widths []int) []string {
	out := make([]string, len(widths))
	for i, w := range widths {
		sep := strings.Repeat("─", w)
		out[i] = ui.Dim(os.Stdout, sep)
	}
	return out
}

func styleCells(headers []string, cells []string, widths []int) []string {
	out := make([]string, len(cells))
	for i, cell := range cells {
		padded := ui.PadRight(cell, widths[i])
		out[i] = ui.StyleCell(fieldLabel(headers[i]), padded)
	}
	return out
}

func humanHeader(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// resolveFields returns the display fields: the provided list if non-empty,
// otherwise all top-level keys from the object.
func resolveFields(obj map[string]interface{}, fields []string) []string {
	if len(fields) > 0 {
		return fields
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	return keys
}

// extractField retrieves a value from obj using a dot-notation path (e.g. "input.title").
func extractField(obj map[string]interface{}, path string) interface{} {
	parts := strings.SplitN(path, ".", 2)
	val, ok := obj[parts[0]]
	if !ok {
		return nil
	}
	if len(parts) == 1 {
		return val
	}
	nested, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	return extractField(nested, parts[1])
}

// fieldLabel returns the last segment of a dot-notation path as the column header.
func fieldLabel(path string) string {
	parts := strings.Split(path, ".")
	return parts[len(parts)-1]
}

// cellString converts a value to a display string, rendering nested objects as compact JSON.
func cellString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case map[string]interface{}, []interface{}:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
