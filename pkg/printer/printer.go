package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const maxCellWidth = 40

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
		// If none of the fields matched directly, recurse into array values
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
		// If any requested field resolves on this object, render it as key-value
		if len(fields) > 0 {
			for _, f := range fields {
				if extractField(v, f) != nil {
					return renderKeyValue(v, fields)
				}
			}
		}
		// Otherwise find the largest array value inside the object
		var best []interface{}
		for _, val := range v {
			if arr, ok := val.([]interface{}); ok && len(arr) > len(best) {
				best = arr
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
		fmt.Println("(no results)")
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	displayHeaders := make([]string, len(headers))
	for i, h := range headers {
		displayHeaders[i] = fieldLabel(h)
	}
	fmt.Fprintln(w, strings.Join(displayHeaders, "\t"))
	fmt.Fprintln(w, strings.Join(repeatStr("---", len(headers)), "\t"))

	for _, row := range rows {
		obj, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		vals := make([]string, len(headers))
		for i, h := range headers {
			vals[i] = truncate(cellString(extractField(obj, h)), maxCellWidth)
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
	return w.Flush()
}

func renderKeyValue(obj map[string]interface{}, fields []string) error {
	keys := resolveFields(obj, fields)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tVALUE")
	fmt.Fprintln(w, "---\t-----")
	for _, k := range keys {
		fmt.Fprintf(w, "%s\t%s\n", fieldLabel(k), truncate(cellString(extractField(obj, k)), maxCellWidth))
	}
	return w.Flush()
}

// resolveFields returns the display fields: the provided list if non-empty,
// otherwise all top-level keys from the object.
func resolveFields(obj map[string]interface{}, fields []string) []string {
	if len(fields) > 0 {
		out := make([]string, 0, len(fields))
		for _, f := range fields {
			if extractField(obj, f) != nil {
				out = append(out, f)
			}
		}
		if len(out) > 0 {
			return out
		}
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

func repeatStr(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}
