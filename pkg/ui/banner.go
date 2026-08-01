package ui

import (
	"fmt"
	"os"
)

const logo = `
   ____                          _
  / ___|_   _ _ __ ___  _   _ __| | ___
 | |  _| | | | '_ ` + "`" + `_ \| | | / _` + "`" + ` |/ _ \
 | |_| | |_| | | | | | | |_| \__, |  __/
  \____|\__,_|_| |_| |_|\__,_|___/\___|`

// PrintLogo renders the Gumlet ASCII logo.
func PrintLogo() {
	if colorsOn(os.Stdout) {
		fmt.Println(Magenta(os.Stdout, logo))
		return
	}
	fmt.Println(logo)
}

// PrintVersionBox shows build metadata in a styled box.
func PrintVersionBox(version, commit, date string) {
	lines := []string{
		fmt.Sprintf("%-8s %s", Dim(os.Stdout, "version"), Bold(os.Stdout, version)),
		fmt.Sprintf("%-8s %s", Dim(os.Stdout, "commit"), commit),
		fmt.Sprintf("%-8s %s", Dim(os.Stdout, "built"), date),
	}
	fmt.Println(Box("gumlet", lines...))
}

// PrintLoginBanner greets the user during login.
func PrintLoginBanner() {
	fmt.Println(Box("Welcome to Gumlet CLI", "Authenticate with your API key to get started."))
}
