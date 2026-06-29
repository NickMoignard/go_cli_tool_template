// Package version exposes build metadata for the CLI.
//
// The unexported vars are injected at release time via
//
//	-ldflags "-X github.com/OWNER/REPLACE_TOOL/internal/version.version=v1.2.3 ..."
//
// (GoReleaser sets these). For plain `go install` builds — where no ldflags are
// passed — Get falls back to runtime/debug.ReadBuildInfo so the reported version,
// commit, and date are still meaningful. See ADR-0004 / the versioning decision.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Injected via -ldflags -X at build time; empty in `go build`/`go install` builds.
var (
	version = ""
	commit  = ""
	date    = ""
)

// Build is the resolved build metadata.
type Build struct {
	Version string
	Commit  string
	Date    string
	Go      string
}

// Get resolves build metadata, preferring ldflags-injected values and falling
// back to the VCS stamps Go records in the build info.
func Get() Build {
	b := Build{Version: version, Commit: commit, Date: date, Go: runtime.Version()}

	if info, ok := debug.ReadBuildInfo(); ok {
		if b.Version == "" && info.Main.Version != "" {
			b.Version = info.Main.Version // e.g. "v1.2.3" or "(devel)"
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if b.Commit == "" {
					b.Commit = s.Value
				}
			case "vcs.time":
				if b.Date == "" {
					b.Date = s.Value
				}
			}
		}
	}

	if b.Version == "" {
		b.Version = "dev"
	}
	if b.Commit == "" {
		b.Commit = "none"
	}
	if b.Date == "" {
		b.Date = "unknown"
	}
	return b
}

// Short returns just the version string, for cobra's rootCmd.Version.
func Short() string { return Get().Version }

// String renders a full, human-readable version line.
func (b Build) String() string {
	return fmt.Sprintf("%s (commit %s, built %s, %s)", b.Version, b.Commit, b.Date, b.Go)
}
