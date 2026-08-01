package ui

import (
	"fmt"
	"io"
	"os"
)

const brandTagline = "Video & image from your terminal"

// PrintBrand renders the Gumlet wordmark with a colored accent bar.
func PrintBrand() {
	accent := Magenta(os.Stdout, "▌")
	title := brandTitle()
	tagline := Dim(os.Stdout, brandTagline)

	fmt.Printf("  %s %s\n", accent, title)
	fmt.Printf("    %s\n", tagline)
}

// PrintLogo renders the Gumlet brand header.
func PrintLogo() {
	PrintBrand()
}

// PrintVersionBox shows build metadata below the brand header.
func PrintVersionBox(version, commit, date string) {
	PrintBrand()
	fmt.Println()
	fmt.Printf("    %-8s %s\n", Dim(os.Stdout, "version"), Bold(os.Stdout, version))
	fmt.Printf("    %-8s %s\n", Dim(os.Stdout, "commit"), commit)
	fmt.Printf("    %-8s %s\n", Dim(os.Stdout, "built"), date)
}

// PrintLoginBanner greets the user during login.
func PrintLoginBanner() {
	PrintBrand()
	fmt.Println()
	fmt.Println("    " + Dim(os.Stdout, "Sign in with your API key to continue."))
}

func brandTitle() string {
	if !colorsOn(os.Stdout) {
		return Bold(os.Stdout, "Gumlet")
	}
	// Gu → magenta, ml → cyan, et → bold white
	parts := []struct {
		text  string
		style func(io.Writer, string) string
	}{
		{"Gu", Magenta},
		{"ml", Cyan},
		{"et", func(w io.Writer, s string) string { return Bold(w, s) }},
	}
	var out string
	for _, p := range parts {
		out += p.style(os.Stdout, p.text)
	}
	return out
}
