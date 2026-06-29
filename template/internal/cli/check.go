package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/OWNER/REPLACE_TOOL/internal/check"
)

// newCheckCmd builds the placeholder `check` subcommand. It reads each input
// (file paths, or stdin via "-" or a pipe), validates it, writes a per-input
// report to stdout (text or JSON), and returns the right exit code: validation
// failures map to ExitInvalid (1), a missing file to ExitUsage (2). Swap the
// check.Check call for real domain logic to build a real tool.
func newCheckCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "check [files...]",
		Short: "Check that inputs are non-empty and valid UTF-8 (placeholder validation)",
		Long: "Check validates each input file (or stdin) against the Template's\n" +
			"placeholder rule: non-empty and valid UTF-8. Pass file paths, use \"-\"\n" +
			"for stdin, or pipe data in. Exit code is 0 if all inputs pass, 1 if any\n" +
			"fail validation, 2 for a usage error (e.g. a missing file).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, opts, args)
		},
	}
}

func runCheck(cmd *cobra.Command, opts *globalOptions, args []string) error {
	log := loggerFrom(cmd.Context())
	stdin := cmd.InOrStdin()
	stdout := cmd.OutOrStdout()
	stderr := cmd.ErrOrStderr()

	names, err := resolveInputNames(args, isTerminal(stdin))
	if err != nil {
		return err
	}

	prog := opts.newProgress(stderr, len(names), "checking")
	results := make([]check.Result, 0, len(names))
	for _, name := range names {
		log.Debug("checking input", "name", display(name))
		r, closeFn, err := openInput(name, stdin)
		if err != nil {
			// A missing/unreadable input is a usage error (ADR-0002).
			return usageErrorf("cannot open %s: %v", display(name), err)
		}
		res := check.Check(display(name), r)
		_ = closeFn()
		results = append(results, res)
		prog.Add(1)
	}
	prog.Finish()

	if err := writeResults(stdout, results, opts.Output); err != nil {
		return internalErr(err)
	}

	failed := 0
	for _, r := range results {
		if !r.OK {
			failed++
		}
	}
	if failed > 0 {
		return validationErrorf("%d of %d inputs failed validation", failed, len(results))
	}
	log.Info("all inputs valid", "count", len(results))
	return nil
}

// resolveInputNames returns the inputs to check; "-" denotes stdin. With no args
// it reads stdin when piped, or errors (usage) when stdin is an interactive TTY.
func resolveInputNames(args []string, stdinIsTTY bool) ([]string, error) {
	if len(args) == 0 {
		if stdinIsTTY {
			return nil, usageErrorf("no input: provide file paths or pipe data to stdin")
		}
		return []string{"-"}, nil
	}
	return args, nil
}

// display is the human-facing name for an input ("-" becomes "<stdin>").
func display(name string) string {
	if name == "-" {
		return "<stdin>"
	}
	return name
}

func openInput(name string, stdin io.Reader) (io.Reader, func() error, error) {
	if name == "-" {
		return stdin, func() error { return nil }, nil
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}

// writeResults renders the report to stdout — the primary, machine-consumable
// output. JSON when requested, otherwise tab-separated human lines.
func writeResults(w io.Writer, results []check.Result, format string) error {
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}
	for _, r := range results {
		if r.OK {
			fmt.Fprintf(w, "ok\t%s\n", r.Name)
		} else {
			fmt.Fprintf(w, "fail\t%s: %s\n", r.Name, r.Problem)
		}
	}
	return nil
}
