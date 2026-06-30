// Package logging builds the CLI's slog.Logger. Logs always go to stderr (the
// data/stderr split, ADR-0002); this package only decides the handler — JSON for
// machines, a colorized console handler (tint) for humans — and the level.
package logging

import (
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

// Format selects the log handler.
type Format string

const (
	FormatAuto Format = "auto" // console on a TTY, JSON otherwise
	FormatText Format = "text" // human-readable console (tint)
	FormatJSON Format = "json" // machine-readable JSON lines
)

// Options configures New. Color and IsTTY are resolved by the caller (it knows
// whether stderr is a terminal and whether --no-color / NO_COLOR applied).
type Options struct {
	Level  slog.Level
	Format Format
	Color  bool
	IsTTY  bool
}

// ParseLevel maps a validated level name ("error"/"warn"/"info"/"debug") to its
// slog.Level. Unknown names fall back to Info; callers validate names upstream.
func ParseLevel(name string) slog.Level {
	switch name {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "debug":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// New returns an *slog.Logger writing to w with the configured handler and level.
// FormatAuto resolves to console output on a TTY and JSON otherwise.
func New(w io.Writer, o Options) *slog.Logger {
	format := o.Format
	if format == FormatAuto {
		if o.IsTTY {
			format = FormatText
		} else {
			format = FormatJSON
		}
	}

	var h slog.Handler
	switch format {
	case FormatJSON:
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: o.Level})
	default: // FormatText: human-readable console via tint.
		h = tint.NewHandler(w, &tint.Options{Level: o.Level, NoColor: !o.Color})
	}
	return slog.New(h)
}
