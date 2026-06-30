// Command yamlvalidate is the generated CLI's entrypoint.
//
// yamlvalidate is a SENTINEL token (ADR-0003): the scaffold renames this
// directory and substitutes the real tool name throughout. Per ADR-0002 the
// entrypoint stays thin — all wiring and the exit-code mapping live in
// internal/cli so they are testable; main only passes os streams and exits.
package main

import (
	"context"
	"os"

	"github.com/NickMoignard/yamlvalidate/internal/cli"
)

func main() {
	os.Exit(cli.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
