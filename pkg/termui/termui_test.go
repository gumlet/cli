package termui

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func TestStylePlainAndColored(t *testing.T) {
	plain := NewStyle(false)
	if got := plain.BoldCyan("TITLE"); got != "TITLE" {
		t.Errorf("plain BoldCyan = %q", got)
	}
	if got := plain.Success("Done."); got != "Done." {
		t.Errorf("plain Success = %q", got)
	}
	if got := plain.Failure("Failed."); got != "Failed." {
		t.Errorf("plain Failure = %q", got)
	}

	colored := NewStyle(true)
	if got := colored.BoldCyan("TITLE"); !strings.Contains(got, "\x1b[") || stripANSI(got) != "TITLE" {
		t.Errorf("colored BoldCyan = %q", got)
	}
	if got := colored.Success("Done."); !strings.Contains(got, "✓") || stripANSI(got) != "✓ Done." {
		t.Errorf("colored Success = %q", got)
	}
	if got := colored.Failure("Failed."); !strings.Contains(got, "✗") || stripANSI(got) != "✗ Failed." {
		t.Errorf("colored Failure = %q", got)
	}
	if got := colored.Hint("More"); !strings.Contains(got, "→") || stripANSI(got) != "→ More" {
		t.Errorf("colored Hint = %q", got)
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestSpinnerNonTTY(t *testing.T) {
	var out bytes.Buffer
	s := startSpinner(&out, "Working...", false, false)
	s.Succeed("All done")
	s.Stop()
	if got := out.String(); got != "Working...\n" {
		t.Errorf("non-TTY spinner = %q, want %q", got, "Working...\n")
	}
}

func TestSpinnerTTY(t *testing.T) {
	var out syncBuffer
	s := startSpinner(&out, "Waiting...", true, true)
	time.Sleep(120 * time.Millisecond)
	s.Succeed("Done!")
	// Calling Stop/Fail after Succeed must be a safe no-op.
	s.Stop()
	s.Fail("Ignored")

	got := out.String()
	if !strings.Contains(got, "Done!") || strings.Contains(got, "Ignored") {
		t.Errorf("unexpected spinner output: %q", got)
	}
}

func TestProgressBar(t *testing.T) {
	var out bytes.Buffer
	bar := newProgressBarWithWidth(&out, "Uploading video", 1024*1024, true, 10)
	bar.Update(512 * 1024)
	bar.Finish("Uploaded video", nil)

	got := out.String()
	if !strings.Contains(got, "50%") || !strings.Contains(got, "100%") || !strings.Contains(got, "Uploaded video") {
		t.Errorf("unexpected progress bar output: %q", got)
	}

	out.Reset()
	barErr := newProgressBarWithWidth(&out, "Uploading video", 1000, true, 10)
	barErr.Update(100)
	barErr.Finish("Uploaded video", errors.New("boom"))
	if !strings.Contains(out.String(), "Uploading video failed") {
		t.Errorf("expected failure message in %q", out.String())
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1536, "1.5 KB"},
		{5 * 1024 * 1024, "5.0 MB"},
		{2 * 1024 * 1024 * 1024, "2.00 GB"},
	}
	for _, tc := range cases {
		if got := formatBytes(tc.in); got != tc.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
