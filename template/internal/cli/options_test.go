package cli

import "testing"

// White-box tests for the flag-resolution logic — the single source of truth for
// the effective log level (ADR/grilling decision: --log-level is canonical,
// --verbose/--debug/--quiet are aliases that adjust it).
func TestGlobalOptions_EffectiveLogLevel(t *testing.T) {
	cases := []struct {
		name string
		o    globalOptions
		want string
	}{
		{"base default", globalOptions{LogLevel: "warn"}, "warn"},
		{"explicit error", globalOptions{LogLevel: "error"}, "error"},
		{"verbose raises warn to info", globalOptions{LogLevel: "warn", Verbose: true}, "info"},
		{"debug raises to debug", globalOptions{LogLevel: "warn", Debug: true}, "debug"},
		{"verbose never lowers an explicit debug", globalOptions{LogLevel: "debug", Verbose: true}, "debug"},
		{"quiet forces error", globalOptions{LogLevel: "warn", Quiet: true}, "error"},
		{"quiet beats verbose and debug", globalOptions{LogLevel: "info", Verbose: true, Debug: true, Quiet: true}, "error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.o.EffectiveLogLevel(); got != tc.want {
				t.Errorf("EffectiveLogLevel() = %q, want %q", got, tc.want)
			}
		})
	}
}
