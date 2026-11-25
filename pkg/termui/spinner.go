package termui

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

const spinnerInterval = 80 * time.Millisecond

// Spinner displays an animated braille spinner on interactive terminals,
// and falls back to a single plain status line on non-TTY writers.
type Spinner struct {
	w        io.Writer
	style    Style
	animated bool
	inner    *spinner.Spinner
	once     sync.Once
}

// StartSpinner starts a spinner with label. On non-TTY writers it writes label
// once as a plain line and does not start a background goroutine.
func StartSpinner(w io.Writer, label string, color bool) *Spinner {
	return startSpinner(w, label, IsTTY(w), color)
}

func startSpinner(w io.Writer, label string, tty, color bool) *Spinner {
	s := &Spinner{
		w:        w,
		style:    NewStyle(color),
		animated: tty,
	}
	if !tty {
		_, _ = fmt.Fprintln(w, label)
		return s
	}

	sp := spinner.New(
		spinner.CharSets[14],
		spinnerInterval,
		spinner.WithWriter(w),
		spinner.WithSuffix(" "+label),
	)
	if color {
		_ = sp.Color("cyan")
	}
	sp.Start()
	s.inner = sp
	return s
}

// Stop halts the spinner and clears its line without printing a final message.
func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	s.finish("")
}

// Succeed halts the spinner and prints a success message on interactive terminals.
func (s *Spinner) Succeed(msg string) {
	if s == nil || !s.animated {
		return
	}
	s.finish(s.style.Success(msg))
}

// Fail halts the spinner and prints a failure message on interactive terminals.
func (s *Spinner) Fail(msg string) {
	if s == nil || !s.animated {
		return
	}
	s.finish(s.style.Failure(msg))
}

func (s *Spinner) finish(finalLine string) {
	s.once.Do(func() {
		if !s.animated || s.inner == nil {
			return
		}
		s.inner.Stop()
		if finalLine != "" {
			_, _ = fmt.Fprintln(s.w, finalLine)
		}
	})
}
