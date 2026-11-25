/*
Copyright © 2026 Gumlet Pte. Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"os"

	"gumlet/cmd"
	"gumlet/pkg/termui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := cmd.NewRootCommand(cmd.Options{Version: version, Commit: commit, Date: date})
	if err := root.Execute(); err != nil {
		stderr := root.ErrOrStderr()
		style := termui.StyleFor(root, stderr)
		if style.Color() {
			fmt.Fprintf(stderr, "%s %v\n", style.BoldRed("✗ Error:"), err)
		} else {
			fmt.Fprintln(stderr, "Error:", err)
		}
		os.Exit(1)
	}
}
