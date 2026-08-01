package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[97m"
)

var colorEnabled = true


// Init configures styling based on flags and environment.
func Init(noColor bool) {
	if noColor || os.Getenv("NO_COLOR") != "" {
		colorEnabled = false
	}
}

func colorsOn(w io.Writer) bool {
	if !colorEnabled {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func paint(w io.Writer, code, s string) string {
	if !colorsOn(w) {
		return s
	}
	return code + s + reset
}

func Bold(w io.Writer, s string) string   { return paint(w, bold, s) }
func Dim(w io.Writer, s string) string    { return paint(w, dim, s) }
func Green(w io.Writer, s string) string  { return paint(w, green, s) }
func Yellow(w io.Writer, s string) string { return paint(w, yellow, s) }
func Red(w io.Writer, s string) string    { return paint(w, red, s) }
func Cyan(w io.Writer, s string) string   { return paint(w, cyan, s) }
func Magenta(w io.Writer, s string) string {
	return paint(w, magenta, s)
}
func Blue(w io.Writer, s string) string { return paint(w, blue, s) }

func Success(msg string, args ...any) {
	line := fmt.Sprintf(msg, args...)
	if colorsOn(os.Stdout) {
		fmt.Fprintf(os.Stdout, "%s %s\n", Green(os.Stdout, "✓"), line)
		return
	}
	fmt.Fprintf(os.Stdout, "OK %s\n", line)
}

func Info(msg string, args ...any) {
	line := fmt.Sprintf(msg, args...)
	if colorsOn(os.Stderr) {
		fmt.Fprintf(os.Stderr, "%s %s\n", Cyan(os.Stderr, "→"), line)
		return
	}
	fmt.Fprintf(os.Stderr, "%s\n", line)
}

func Hint(msg string, args ...any) {
	line := fmt.Sprintf(msg, args...)
	if colorsOn(os.Stdout) {
		fmt.Fprintf(os.Stdout, "  %s %s\n", Dim(os.Stdout, "tip:"), Dim(os.Stdout, line))
		return
	}
	fmt.Fprintf(os.Stdout, "tip: %s\n", line)
}

func Warn(msg string, args ...any) {
	line := fmt.Sprintf(msg, args...)
	if colorsOn(os.Stderr) {
		fmt.Fprintf(os.Stderr, "%s %s\n", Yellow(os.Stderr, "!"), line)
		return
	}
	fmt.Fprintf(os.Stderr, "WARN %s\n", line)
}

func Error(msg string, args ...any) {
	line := fmt.Sprintf(msg, args...)
	if colorsOn(os.Stderr) {
		fmt.Fprintf(os.Stderr, "%s %s\n", Red(os.Stderr, "✗"), line)
		return
	}
	fmt.Fprintf(os.Stderr, "ERROR %s\n", line)
}

func Prompt(label string) string {
	if colorsOn(os.Stdout) {
		return Cyan(os.Stdout, label)
	}
	return label
}

// Box draws a simple titled box around lines.
func Box(title string, lines ...string) string {
	titleLine := Bold(os.Stdout, title)
	content := append([]string{titleLine}, lines...)

	width := visibleLen(titleLine)
	for _, line := range lines {
		if w := visibleLen(line); w > width {
			width = w
		}
	}
	if width < 24 {
		width = 24
	}

	var b strings.Builder
	top := "╭" + strings.Repeat("─", width+2) + "╮"
	bot := "╰" + strings.Repeat("─", width+2) + "╯"
	b.WriteString(top + "\n")
	for _, line := range content {
		padding := width - visibleLen(line)
		if padding < 0 {
			padding = 0
		}
		b.WriteString(fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding)))
	}
	b.WriteString(bot)
	return b.String()
}

func visibleLen(s string) int {
	// Strip ANSI sequences for width calculation.
	out := s
	for {
		start := strings.Index(out, "\033[")
		if start < 0 {
			break
		}
		end := strings.Index(out[start:], "m")
		if end < 0 {
			break
		}
		out = out[:start] + out[start+end+1:]
	}
	return len(out)
}

// StyleCell applies semantic coloring for known field values in tables.
// value should already be padded to the target column width.
func StyleCell(header, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "—" {
		return Dim(os.Stdout, value)
	}

	switch strings.ToLower(header) {
	case "status":
		return styleStatus(value)
	case "is_active", "active":
		return styleBool(value)
	default:
		return value
	}
}

// PadRight pads plain text to width (no ANSI).
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func styleStatus(status string) string {
	switch strings.ToLower(status) {
	case "ready", "completed", "active", "success":
		return Green(os.Stdout, status)
	case "processing", "upload-pending", "pending", "queued":
		return Yellow(os.Stdout, status)
	case "failed", "error", "deleted", "inactive":
		return Red(os.Stdout, status)
	default:
		return status
	}
}

func styleBool(value string) string {
	switch strings.ToLower(value) {
	case "true", "yes":
		return Green(os.Stdout, value)
	case "false", "no":
		return Dim(os.Stdout, value)
	default:
		return value
	}
}

// CountLabel returns a friendly footer for result sets.
func CountLabel(n int, singular, plural string) string {
	word := plural
	if n == 1 {
		word = singular
	}
	if colorsOn(os.Stdout) {
		return Dim(os.Stdout, fmt.Sprintf("%d %s", n, word))
	}
	return fmt.Sprintf("%d %s", n, word)
}

// EmptyState message when a list has no rows.
func EmptyState() string {
	if colorsOn(os.Stdout) {
		return Dim(os.Stdout, "No results found — try adjusting your filters?")
	}
	return "No results found."
}
