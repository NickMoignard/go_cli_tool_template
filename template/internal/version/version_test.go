package version

import (
	"strings"
	"testing"
)

// In a `go test` binary no -ldflags are injected, so Get exercises the
// ReadBuildInfo fallback and then the hardcoded sentinels. The behavioral
// guarantee we care about: --version is never blank, so no field is ever empty.
func TestGet_NeverReturnsEmptyFields(t *testing.T) {
	b := Get()

	if b.Version == "" {
		t.Error("Version is empty; want a non-empty fallback (e.g. \"dev\")")
	}
	if b.Commit == "" {
		t.Error("Commit is empty; want a non-empty fallback (e.g. \"none\")")
	}
	if b.Date == "" {
		t.Error("Date is empty; want a non-empty fallback (e.g. \"unknown\")")
	}
	if b.Go == "" {
		t.Error("Go is empty; want runtime.Version()")
	}
}

func TestBuildString_IncludesAllFields(t *testing.T) {
	b := Build{Version: "v1.2.3", Commit: "abc1234", Date: "2026-06-29T00:00:00Z", Go: "go1.26"}

	got := b.String()
	for _, want := range []string{b.Version, b.Commit, b.Date, b.Go} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}
