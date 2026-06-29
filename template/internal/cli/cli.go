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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/OWNER/REPLACE_TOOL/internal/config"
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

// usageError marks a problem with how the command was invoked (bad flag value,
// wrong args). It maps to ExitUsage (2).
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }
func (e *usageError) ExitCode() int { return ExitUsage }

func usageErrorf(format string, a ...any) error {
	return &usageError{msg: fmt.Sprintf(format, a...)}
}

// validationError marks input that the tool processed correctly but found
// non-conforming (e.g. a YAML file that fails its schema). It maps to
// ExitInvalid (1) — distinct from a tool failure — so scripts can tell
// "the input is wrong" from "the tool broke" (ADR-0002).
type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) ExitCode() int { return ExitInvalid }

func validationErrorf(format string, a ...any) error {
	return &validationError{msg: fmt.Sprintf(format, a...)}
}

// internalError wraps an unexpected failure. It maps to ExitInternal (>2) and
// preserves the cause via Unwrap.
type internalError struct{ err error }

func (e *internalError) Error() string { return e.err.Error() }
func (e *internalError) ExitCode() int { return ExitInternal }
func (e *internalError) Unwrap() error { return e.err }

func internalErr(err error) error { return &internalError{err: err} }

// globalOptions holds the resolved values of the persistent (global) flags.
// Flags bind into it; downstream beads consume it (output rendering F7, logging
// F4, config F6). Keeping it a struct gives those a single, testable source.
type globalOptions struct {
	Output    string // "text" | "json"
	LogLevel  string // "error" | "warn" | "info" | "debug" (base level)
	LogFormat string // "auto" | "text" | "json"
	Verbose   bool   // -v: raise effective level to at least info
	Debug     bool   // --debug: raise effective level to debug
	Quiet     bool   // -q: force effective level to error (and suppress progress, F5)
	NoColor   bool   // --no-color: disable ANSI (also honors NO_COLOR / non-TTY, F4)
	NoInput   bool   // --no-input: never prompt; assume non-interactive
	Config    string // --config: explicit config file path (loaded in F6)
}

// logLevels orders levels from least to most verbose; index is the rank.
var logLevels = []string{"error", "warn", "info", "debug"}

func logLevelRank(name string) int {
	for i, l := range logLevels {
		if l == name {
			return i
		}
	}
	return 0 // unreachable after validate(); default to the quietest
}

// EffectiveLogLevel resolves the single log level from the base --log-level plus
// the alias flags. Rule: start at --log-level; --verbose raises to at least info
// and --debug to at least debug (the most verbose request wins); --quiet then
// overrides everything to error. The logger (F4) consumes this.
func (o *globalOptions) EffectiveLogLevel() string {
	rank := logLevelRank(o.LogLevel)
	if o.Verbose && rank < logLevelRank("info") {
		rank = logLevelRank("info")
	}
	if o.Debug && rank < logLevelRank("debug") {
		rank = logLevelRank("debug")
	}
	if o.Quiet {
		rank = logLevelRank("error")
	}
	return logLevels[rank]
}

// applyConfig overlays loaded config (env > file > defaults) onto the options,
// but only for flags the user did not explicitly set — preserving the top of the
// precedence chain: flags > env > file > defaults.
func (o *globalOptions) applyConfig(cfg config.Config, changed func(name string) bool) {
	if !changed("output") {
		o.Output = cfg.Output
	}
	if !changed("log-level") {
		o.LogLevel = cfg.LogLevel
	}
	if !changed("log-format") {
		o.LogFormat = cfg.LogFormat
	}
	if !changed("no-color") {
		o.NoColor = cfg.NoColor
	}
}

// configSearchDirs lists where a config file is discovered when --config is not
// given: the current directory (project-local .<app>.yaml) and the user config
// dir ($XDG_CONFIG_HOME/<app>/config.yaml).
func configSearchDirs() []string {
	dirs := []string{"."}
	if d, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(d, binaryName))
	}
	return dirs
}

// validate checks the resolved flag values, returning a usageError (exit 2) for
// any invalid combination. Called from the root PersistentPreRunE so it runs for
// every (sub)command.
func (o *globalOptions) validate() error {
	switch o.Output {
	case "text", "json":
	default:
		return usageErrorf("invalid --output %q: want %q or %q", o.Output, "text", "json")
	}
	switch o.LogLevel {
	case "error", "warn", "info", "debug":
	default:
		return usageErrorf("invalid --log-level %q: want one of error, warn, info, debug", o.LogLevel)
	}
	switch o.LogFormat {
	case "auto", "text", "json":
	default:
		return usageErrorf("invalid --log-format %q: want auto, text, or json", o.LogFormat)
	}
	return nil
}

// NewRootCmd builds the root command and binds the global flag set. Subcommands
// (F7) are attached by later work; today the bare command prints help.
func NewRootCmd() *cobra.Command {
	opts := &globalOptions{}

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load config (env > file > defaults), then let explicitly-set flags
			// win over it before validating the merged result.
			cfg, err := config.Load(config.Source{
				AppName:      binaryName,
				ExplicitPath: opts.Config,
				SearchDirs:   configSearchDirs(),
				Environ:      os.Environ(),
			})
			if err != nil {
				return usageErrorf("%v", err)
			}
			opts.applyConfig(cfg, cmd.Flags().Changed)

			if err := opts.validate(); err != nil {
				return err
			}
			// Build the logger from resolved flags and stash it on the command's
			// context for subcommands to retrieve via loggerFrom. NOTE: cobra runs
			// only the nearest PersistentPreRunE; a subcommand that defines its own
			// must re-invoke this wiring (or leave it unset).
			logger := opts.newLogger(cmd.ErrOrStderr())
			cmd.SetContext(context.WithValue(cmd.Context(), loggerKey, logger))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// No subcommand yet: show help. F7 replaces this with real behavior.
			return cmd.Help()
		},
	}
	cmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")

	flags := cmd.PersistentFlags()
	flags.StringVarP(&opts.Output, "output", "o", "text", "output format: text or json")
	flags.StringVar(&opts.LogLevel, "log-level", "warn", "log verbosity: error, warn, info, or debug")
	flags.StringVar(&opts.LogFormat, "log-format", "auto", "log format: auto, text, or json")
	// Registering -v for verbose before Execute makes cobra add --version without
	// the -v shorthand (it only claims -v when free), reclaiming -v as we want.
	flags.BoolVarP(&opts.Verbose, "verbose", "v", false, "raise log verbosity to info (alias for --log-level info)")
	flags.BoolVar(&opts.Debug, "debug", false, "raise log verbosity to debug (alias for --log-level debug)")
	flags.BoolVarP(&opts.Quiet, "quiet", "q", false, "suppress progress and non-error logs")
	flags.BoolVar(&opts.NoColor, "no-color", false, "disable colored output")
	flags.BoolVar(&opts.NoInput, "no-input", false, "never prompt; run non-interactively")
	flags.StringVar(&opts.Config, "config", "", "path to a config file")

	return cmd
}

// Run executes the command tree against the given streams and returns the process
// exit code. main() is a one-liner around this.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return runCmd(ctx, NewRootCmd(), args, stdin, stdout, stderr)
}

// runCmd runs a prepared root command and maps its result to an exit code. It is
// separate from Run so tests can supply a root with extra subcommands.
func runCmd(ctx context.Context, root *cobra.Command, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
