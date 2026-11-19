package printer

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

// JoinStrings formats an explicitly selected string or array of strings.
func JoinStrings(separator string) func(any, map[string]any) string {
	return func(value any, _ map[string]any) string {
		if text, ok := value.(string); ok {
			return text
		}
		values, _ := value.([]any)
		parts := make([]string, 0, len(values))
		for _, entry := range values {
			if text, ok := entry.(string); ok && text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, separator)
	}
}

// Date formats an RFC 3339 string or a Unix timestamp as a UTC date or time.
func Date(layout string, fromUnix func(int64) time.Time) func(any, map[string]any) string {
	return func(value any, _ map[string]any) string {
		var t time.Time
		switch v := value.(type) {
		case json.Number:
			n, err := v.Int64()
			if err != nil || fromUnix == nil {
				return ""
			}
			t = fromUnix(n)
		case string:
			var err error
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return v
			}
		default:
			return ""
		}
		return t.UTC().Format(layout)
	}
}

const UTCDateTime = "2006-01-02 15:04 UTC"

// PublicLink selects the first unsigned HTTP(S) link from a string or array.
// Raw URLs that may carry credentials remain available via JSON output.
func PublicLink(value any, _ map[string]any) string {
	if list, ok := value.([]any); ok {
		for _, item := range list {
			if link := publicLink(item); link != "" {
				return link
			}
		}
		return ""
	}
	return publicLink(value)
}

func publicLink(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	u, err := url.Parse(text)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") ||
		u.User != nil || strings.ContainsAny(text, "?#") {
		return ""
	}
	return text
}
