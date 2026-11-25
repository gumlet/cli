package termui

import (
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// IsTTY reports whether w is an interactive terminal capable of ANSI control sequences.
func IsTTY(w io.Writer) bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}

// ColorEnabled reports whether styled/colored output should be used for w.
// It honors --no-color, the NO_COLOR environment variable, TERM=dumb, and TTY detection.
func ColorEnabled(cmd *cobra.Command, w io.Writer) bool {
	if cmd != nil {
		if disabled, _ := cmd.Flags().GetBool("no-color"); disabled {
			return false
		}
	}
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		return false
	}
	return IsTTY(w)
}

// Style formats terminal text using lipgloss when color is enabled,
// and returns plain text when color is disabled.
type Style struct {
	color     bool
	bold      lipgloss.Style
	dim       lipgloss.Style
	red       lipgloss.Style
	green     lipgloss.Style
	yellow    lipgloss.Style
	cyan      lipgloss.Style
	boldRed   lipgloss.Style
	boldGreen lipgloss.Style
	boldCyan  lipgloss.Style
	link      lipgloss.Style
}

// NewStyle creates a Style with explicit color enablement.
func NewStyle(color bool) Style {
	if !color {
		return Style{}
	}
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.ANSI)
	return Style{
		color:     true,
		bold:      r.NewStyle().Bold(true),
		dim:       r.NewStyle().Faint(true),
		red:       r.NewStyle().Foreground(lipgloss.Color("1")),
		green:     r.NewStyle().Foreground(lipgloss.Color("2")),
		yellow:    r.NewStyle().Foreground(lipgloss.Color("3")),
		cyan:      r.NewStyle().Foreground(lipgloss.Color("6")),
		boldRed:   r.NewStyle().Bold(true).Foreground(lipgloss.Color("1")),
		boldGreen: r.NewStyle().Bold(true).Foreground(lipgloss.Color("2")),
		boldCyan:  r.NewStyle().Bold(true).Foreground(lipgloss.Color("6")),
		link:      r.NewStyle().Underline(true).Foreground(lipgloss.Color("6")),
	}
}

// StyleFor creates a Style based on command flags, environment, and writer TTY state.
func StyleFor(cmd *cobra.Command, w io.Writer) Style {
	return NewStyle(ColorEnabled(cmd, w))
}

// Color reports whether ANSI colors are enabled.
func (s Style) Color() bool { return s.color }

func (s Style) render(st lipgloss.Style, text string) string {
	if !s.color || text == "" {
		return text
	}
	return st.Render(text)
}

func (s Style) Bold(text string) string      { return s.render(s.bold, text) }
func (s Style) Dim(text string) string       { return s.render(s.dim, text) }
func (s Style) Red(text string) string       { return s.render(s.red, text) }
func (s Style) Green(text string) string     { return s.render(s.green, text) }
func (s Style) Yellow(text string) string    { return s.render(s.yellow, text) }
func (s Style) Cyan(text string) string      { return s.render(s.cyan, text) }
func (s Style) BoldRed(text string) string   { return s.render(s.boldRed, text) }
func (s Style) BoldGreen(text string) string { return s.render(s.boldGreen, text) }
func (s Style) BoldCyan(text string) string  { return s.render(s.boldCyan, text) }
func (s Style) Link(text string) string      { return s.render(s.link, text) }

func (s Style) Success(msg string) string {
	if !s.color {
		return msg
	}
	return s.BoldGreen("✓") + " " + msg
}

func (s Style) Failure(msg string) string {
	if !s.color {
		return msg
	}
	return s.BoldRed("✗") + " " + msg
}

func (s Style) Warn(msg string) string {
	if !s.color {
		return msg
	}
	return s.Yellow("!") + " " + msg
}

func (s Style) Hint(msg string) string {
	if !s.color {
		return msg
	}
	return s.Cyan("→") + " " + s.Dim(msg)
}
