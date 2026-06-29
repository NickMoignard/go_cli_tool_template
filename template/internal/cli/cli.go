// Package cli wires the command tree and translates results into process exit
// codes. Keeping the mapping here (rather than in main) makes it unit-testable.
//
// Exit-code contract (ADR-0002):
//
//	0  success / valid
//	1  validation failure (ran fine, input did not conform)
//	2  usage error (bad flags, missing/unknown command or args)
//	>2 internal / unexpected error
//
// Convention: domain code signals 1 or an internal code by returning a
// CodedError. Any *uncoded* error reaching Run came from cobra's flag/arg/command
// parsing and is therefore a usage error (2). Later work (F3) refines the typed
// error set and adds the global flag set; F7 adds the real subcommand.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/OWNER/REPLACE_TOOL/internal/version"
)

// Exit codes per ADR-0002.
const (
	ExitOK       = 0
	ExitInvalid  = 1
	ExitUsage    = 2
	ExitInternal = 3
)

// binaryName is a sentinel token replaced by the scaffold (ADR-0003).
const binaryName = "REPLACE_TOOL"

// CodedError lets domain code attach a specific exit code to an error. Validation
// failures return a CodedError with ExitInvalid; unexpected conditions use a code
// > ExitUsage. F3 provides concrete typed errors implementing this.
type CodedError interface {
	error
	ExitCode() int
}

// NewRootCmd builds the root command. Global flags (F3) and subcommands (F7) are
// attached by later work; today the bare command prints help and serves --version.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     binaryName,
		Short:   binaryName + " — a CLI scaffolded from go_cli_tool_template",
		Version: version.Short(),
		// No subcommands yet, so reject stray positional args — an unknown
		// command/arg must be a usage error (2), not silently swallowed. F7
		// attaches real subcommands and relaxes this as appropriate.
		Args: cobra.NoArgs,
		// We own error rendering and the usage-vs-error decision (ADR-0002), so
		// silence cobra's built-in printing.
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// No subcommand yet: show help. F7 replaces this with real behavior.
			return cmd.Help()
		},
	}
	cmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	return cmd
}

// Run executes the command tree against the given streams and returns the process
// exit code. main() is a one-liner around this.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root := NewRootCmd()
	root.SetArgs(args)
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)

	err := root.ExecuteContext(ctx)
	if err == nil {
		return ExitOK
	}

	fmt.Fprintln(stderr, "error:", err)

	var coded CodedError
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	// Uncoded error => cobra parse/dispatch problem => usage error.
	return ExitUsage
}
