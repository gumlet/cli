package termui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	defaultBarWidth = 24
	progressTick    = 50 * time.Millisecond
)

// ProgressBar renders a smooth single-line transfer progress bar on interactive terminals
// using schollz/progressbar.
type ProgressBar struct {
	w     io.Writer
	label string
	total int64
	style Style
	bar   *progressbar.ProgressBar
	width int
}

// NewProgressBar returns a ProgressBar if w is an interactive terminal and total > 0.
// Otherwise it returns nil so callers can skip progress rendering cleanly.
func NewProgressBar(w io.Writer, label string, total int64, color bool) *ProgressBar {
	if !IsTTY(w) || total <= 0 {
		return nil
	}
	return newProgressBarWithWidth(w, label, total, color, defaultBarWidth)
}

func newProgressBarWithWidth(w io.Writer, label string, total int64, color bool, width int) *ProgressBar {
	style := NewStyle(color)
	theme := progressbar.Theme{
		Saucer:        style.Cyan("━"),
		SaucerHead:    style.Cyan("━"),
		SaucerPadding: style.Dim("─"),
		BarStart:      "",
		BarEnd:        "",
	}
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetWriter(w),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription(label),
		progressbar.OptionSetWidth(width),
		progressbar.OptionThrottle(progressTick),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(theme),
		progressbar.OptionShowCount(),
	)
	return &ProgressBar{
		w:     w,
		label: label,
		total: total,
		style: style,
		bar:   bar,
		width: width,
	}
}

// Update advances the progress bar to current bytes transferred.
func (p *ProgressBar) Update(current int64) {
	if p == nil {
		return
	}
	if current > p.total {
		current = p.total
	}
	_ = p.bar.Set64(current)
}

// Finish finalizes the progress bar line and writes a crisp completion summary.
func (p *ProgressBar) Finish(doneLabel string, err error) {
	if p == nil {
		return
	}
	_ = p.bar.Clear()
	if err != nil {
		_, _ = fmt.Fprintln(p.w, p.style.Failure(p.label+" failed"))
		return
	}
	bar := p.style.Green(strings.Repeat("━", p.width))
	_, _ = fmt.Fprintf(p.w, "%s  %s %s  %s\n",
		p.style.Success(doneLabel),
		bar,
		p.style.BoldGreen("100%"),
		p.style.Dim(formatBytes(p.total)),
	)
}

func formatBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
