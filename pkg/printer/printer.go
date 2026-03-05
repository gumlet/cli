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
		v = filterValue(v, fields)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(string(out))
	return nil
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
		// Find the largest array value inside the object
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
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Join(repeatStr("---", len(headers)), "\t"))

	for _, row := range rows {
		obj, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		vals := make([]string, len(headers))
		for i, h := range headers {
			vals[i] = truncate(cellString(obj[h]), maxCellWidth)
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
		fmt.Fprintf(w, "%s\t%s\n", k, truncate(cellString(obj[k]), maxCellWidth))
	}
	return w.Flush()
}

// resolveFields returns the display fields: the provided list if non-empty,
// otherwise all keys from the object.
func resolveFields(obj map[string]interface{}, fields []string) []string {
	if len(fields) > 0 {
		// Only include fields that actually exist in the object
		out := make([]string, 0, len(fields))
		for _, f := range fields {
			if _, ok := obj[f]; ok {
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

// filterValue filters a JSON-decoded value to only include the given fields.
func filterValue(v interface{}, fields []string) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// Check if any of the fields exist directly; if so, filter this object
		for _, f := range fields {
			if _, ok := val[f]; ok {
				out := make(map[string]interface{}, len(fields))
				for _, f2 := range fields {
					if fv, ok := val[f2]; ok {
						out[f2] = fv
					}
				}
				return out
			}
		}
		// Fields not at this level — recurse into values to find arrays
		for k, child := range val {
			val[k] = filterValue(child, fields)
		}
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = filterValue(item, fields)
		}
		return val
	}
	return v
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
