package cli_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/OWNER/REPLACE_TOOL/internal/cli"
)

// TestMain lets testscript re-invoke this binary in-process as the command
// "REPLACE_TOOL", so .txtar scripts exercise the real CLI end-to-end (real
// streams, real exit status) rather than calling cli.Run directly.
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"REPLACE_TOOL": func() int {
			return cli.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
		},
	}))
}

// TestScripts runs every testdata/script/*.txtar end-to-end. testscript asserts
// success/failure and stream contents; exact exit-code values (0/1/2/>2) are
// pinned by the cli.Run buffer tests in cli_test.go.
func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "script"),
	})
}
