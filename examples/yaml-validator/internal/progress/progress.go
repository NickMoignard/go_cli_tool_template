// Package progress reports progress of multi-step operations. Output always goes
// to the writer it is built with (stderr in practice, never stdout — progress
// must not corrupt piped data, ADR-0002 / clig.dev). When disabled (non-TTY,
// --quiet, or --no-input) it degrades to a silent no-op, so call sites stay clean.
package progress

import (
	"io"

	"github.com/schollz/progressbar/v3"
)

// Tracker advances a progress indicator. Add reports completed units; Finish
// completes it. A Tracker is always non-nil, so callers never nil-check.
type Tracker interface {
	Add(n int)
	Finish()
}

// Options configures New. Enabled is resolved by the caller (TTY and neither
// --quiet nor --no-input). Total <= 0 means an indeterminate amount of work.
type Options struct {
	Enabled     bool
	Total       int
	Description string
}

// New returns a Tracker writing to w, or a silent no-op when Options.Enabled is
// false.
func New(w io.Writer, o Options) Tracker {
	if !o.Enabled {
		return noop{}
	}
	return &bar{pb: progressbar.NewOptions(o.Total,
		progressbar.OptionSetWriter(w),
		progressbar.OptionSetDescription(o.Description),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionShowCount(),
	)}
}

type noop struct{}

func (noop) Add(int) {}
func (noop) Finish() {}

type bar struct{ pb *progressbar.ProgressBar }

func (b *bar) Add(n int) { _ = b.pb.Add(n) }
func (b *bar) Finish()   { _ = b.pb.Finish() }
