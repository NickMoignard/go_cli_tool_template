package cli

import (
	"context"
	"io"
	"log/slog"
	"os"

	"golang.org/x/term"

	"github.com/OWNER/REPLACE_TOOL/internal/logging"
)

// ctxKey is an unexported context key type to avoid collisions.
type ctxKey int

const loggerKey ctxKey = iota

// loggerFrom returns the logger stashed in ctx by the root PersistentPreRunE, or
// a discard logger if none is present (e.g. unit tests that bypass wiring). It
// never returns nil, so callers can log unconditionally.
func loggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newLogger builds the logger from the resolved options, writing to stderr. It
// resolves TTY-ness (for FormatAuto) and whether color applies (TTY and neither
// --no-color nor NO_COLOR disabled it).
func (o *globalOptions) newLogger(stderr io.Writer) *slog.Logger {
	isTTY := isTerminal(stderr)
	color := isTTY && !o.NoColor && os.Getenv("NO_COLOR") == ""
	return logging.New(stderr, logging.Options{
		Level:  logging.ParseLevel(o.EffectiveLogLevel()),
		Format: logging.Format(o.LogFormat),
		Color:  color,
		IsTTY:  isTTY,
	})
}

// isTerminal reports whether w is a terminal (so non-file writers like test
// buffers are treated as non-TTY).
func isTerminal(w io.Writer) bool {
	f, ok := w.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(int(f.Fd()))
}
