package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/NickMoignard/yamlvalidate/internal/validate"
)

// newValidateCmd builds the `validate` subcommand: it validates each input
// (file paths, or stdin via "-" or a pipe) against the JSON Schema named by
// --schema, writes a per-input report to stdout (text or JSON), and returns the
// exit code from the contract — 0 if every document conforms, 1 if any fails
// validation, 2 for a usage error (bad/missing schema or input file).
func newValidateCmd(opts *globalOptions) *cobra.Command {
	var schemaPath string
	cmd := &cobra.Command{
		Use:   "validate [files...]",
		Short: "Validate YAML files against a JSON Schema",
		Long: "Validate checks each input file (or stdin) against the JSON Schema given\n" +
			"by --schema (Draft 2020-12). Pass file paths, use \"-\" for stdin, or pipe\n" +
			"data in. Failures are reported as path-addressed violations. Exit code is 0\n" +
			"if all inputs conform, 1 if any fail validation, 2 for a usage error (e.g. a\n" +
			"missing schema or input file).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd, opts, schemaPath, args)
		},
	}
	cmd.Flags().StringVarP(&schemaPath, "schema", "s", "", "path to the JSON Schema to validate against (required)")
	return cmd
}

func runValidate(cmd *cobra.Command, opts *globalOptions, schemaPath string, args []string) error {
	log := loggerFrom(cmd.Context())
	stdin := cmd.InOrStdin()
	stdout := cmd.OutOrStdout()
	stderr := cmd.ErrOrStderr()

	if schemaPath == "" {
		return usageErrorf("no schema: pass --schema <file>")
	}
	validator, err := validate.NewValidator(schemaPath)
	if err != nil {
		// A bad or unreadable schema is a usage error (ADR-0002).
		return usageErrorf("%v", err)
	}

	names, err := resolveInputNames(args, isTerminal(stdin))
	if err != nil {
		return err
	}

	prog := opts.newProgress(stderr, len(names), "validating")
	results := make([]validate.Result, 0, len(names))
	for _, name := range names {
		log.Debug("validating input", "name", display(name))
		r, closeFn, err := openInput(name, stdin)
		if err != nil {
			return usageErrorf("cannot open %s: %v", display(name), err)
		}
		res := validator.Validate(display(name), r)
		_ = closeFn()
		results = append(results, res)
		prog.Add(1)
	}
	prog.Finish()

	if err := writeValidateResults(stdout, results, opts.Output); err != nil {
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

// resolveInputNames returns the inputs to process; "-" denotes stdin. With no
// args it reads stdin when piped, or errors (usage) when stdin is an interactive
// TTY.
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

// writeValidateResults renders the report to stdout — the primary,
// machine-consumable output. JSON when requested, otherwise human lines with one
// indented line per violation.
func writeValidateResults(w io.Writer, results []validate.Result, format string) error {
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}
	for _, r := range results {
		if r.OK {
			fmt.Fprintf(w, "ok\t%s\n", r.Name)
			continue
		}
		fmt.Fprintf(w, "fail\t%s\n", r.Name)
		for _, v := range r.Violations {
			path := v.Path
			if path == "" {
				path = "/"
			}
			fmt.Fprintf(w, "\t%s: %s\n", path, v.Message)
		}
	}
	return nil
}
