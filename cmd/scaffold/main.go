// Command scaffold instantiates the ./template module into a brand-new Go CLI
// project: it copies template/, replaces the sentinel tokens, renames the
// cmd/<tool> directory, and runs `go mod tidy` + `git init` (ADR-0003).
//
// All logic lives in internal/scaffold so it stays testable; main only wires
// os streams and the exit code.
package main

import (
	"os"

	"github.com/NickMoignard/go_cli_tool_template/internal/scaffold"
)

func main() {
	os.Exit(scaffold.Run(os.Args[1:], os.Stdout, os.Stderr))
}
