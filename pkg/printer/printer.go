package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// Print outputs data in the specified format ("json" or "table").
func Print(data []byte, format string) error {
	switch format {
	case "table":
		return printTable(data)
	default:
		return printJSON(data)
	}
}

func printJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return nil
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(string(out))
	return nil
}

func printTable(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Println(string(data))
		return nil
	}

	switch v := raw.(type) {
	case []interface{}:
		return renderSlice(v)
	case map[string]interface{}:
		// Look for an array value inside the object
		for _, val := range v {
			if arr, ok := val.([]interface{}); ok {
				return renderSlice(arr)
			}
		}
		// No array found: render as key-value
		return renderKeyValue(v)
	default:
		fmt.Println(string(data))
	}
	return nil
}

func renderSlice(rows []interface{}) error {
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

	headers := orderedKeys(firstObj)

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
			vals[i] = fmt.Sprintf("%v", obj[h])
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
	return w.Flush()
}

func renderKeyValue(obj map[string]interface{}) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tVALUE")
	fmt.Fprintln(w, "---\t---")
	for k, v := range obj {
		fmt.Fprintf(w, "%s\t%v\n", k, v)
	}
	return w.Flush()
}

func orderedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func repeatStr(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}
