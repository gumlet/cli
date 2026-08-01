package ui

import (
	"fmt"
	"os"
	"strings"
)

const barWidth = 28

// ProgressBar renders an upload/download progress line to stderr.
func ProgressBar(label string, read, total int64) {
	if total <= 0 {
		return
	}

	pct := float64(read) / float64(total)
	filled := int(pct * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	pctStr := fmt.Sprintf("%5.1f%%", pct*100)
	sizeStr := fmt.Sprintf("%s / %s", formatBytes(read), formatBytes(total))

	if colorsOn(os.Stderr) {
		fmt.Fprintf(os.Stderr, "\r%s %s %s  %s",
			Cyan(os.Stderr, label),
			Green(os.Stderr, bar),
			Dim(os.Stderr, pctStr),
			sizeStr,
		)
		return
	}
	fmt.Fprintf(os.Stderr, "\r%s [%s] %s %s", label, bar, pctStr, sizeStr)
}

// FinishProgress clears the progress line.
func FinishProgress() {
	fmt.Fprintln(os.Stderr)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
