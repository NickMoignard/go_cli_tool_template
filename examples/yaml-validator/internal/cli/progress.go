package cli

import (
	"io"

	"github.com/NickMoignard/yamlvalidate/internal/progress"
)

// progressEnabled decides whether to show progress: only on a TTY, and never when
// the user asked for quiet or non-interactive operation (clig.dev / ADR-0002).
func progressEnabled(isTTY, quiet, noInput bool) bool {
	return isTTY && !quiet && !noInput
}

// newProgress builds a progress tracker writing to stderr, automatically silent
// when stderr is not a terminal or --quiet/--no-input is set. Subcommands call
// this to report progress without worrying about when it is appropriate.
func (o *globalOptions) newProgress(stderr io.Writer, total int, description string) progress.Tracker {
	return progress.New(stderr, progress.Options{
		Enabled:     progressEnabled(isTerminal(stderr), o.Quiet, o.NoInput),
		Total:       total,
		Description: description,
	})
}
