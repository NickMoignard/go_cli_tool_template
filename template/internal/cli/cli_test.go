package cli_test

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OWNER/REPLACE_TOOL/internal/cli"
)

var update = flag.Bool("update", false, "update golden files")

// assertGolden compares actual against testdata/<golden>, or rewrites it when
// -update is passed. Golden files keep assertions on rich, multi-line output
// (like --help) maintainable as the command surface grows.
func assertGolden(t *testing.T, golden, actual string) {
	t.Helper()
	path := filepath.Join("testdata", golden)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(actual), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s: %v (run `go test -update` to create)", path, err)
	}
	if actual != string(want) {
		t.Errorf("output does not match %s:\n--- got ---\n%s\n--- want ---\n%s", path, actual, want)
	}
}

// run is a small helper: it executes cli.Run with buffered streams and returns
// the exit code plus whatever landed on stdout and stderr. Testing through Run
// (the public entrypoint) keeps these tests behavioral — they assert the
// stdout-vs-stderr contract and exit codes, not internal wiring.
func run(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = cli.Run(context.Background(), args, strings.NewReader(""), &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func TestRun_Version_WritesVersionToStdout(t *testing.T) {
	code, stdout, stderr := run(t, "--version")

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK)", code, cli.ExitOK)
	}
	if !strings.Contains(stdout, "REPLACE_TOOL") {
		t.Errorf("stdout = %q, want it to contain the binary name", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty (version is data → stdout)", stderr)
	}
}

func TestRun_Help_MatchesGoldenOnStdout(t *testing.T) {
	code, stdout, stderr := run(t, "--help")

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK)", code, cli.ExitOK)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty (requested help is output → stdout)", stderr)
	}
	assertGolden(t, "help.golden", stdout)
}

func TestRun_OutputFlag_AcceptsJSON(t *testing.T) {
	code, _, stderr := run(t, "-o", "json")

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
}

func TestRun_OutputFlag_RejectsInvalidValue(t *testing.T) {
	code, stdout, stderr := run(t, "-o", "xml")

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty (validation failure renders to stderr)", stdout)
	}
	if !strings.Contains(stderr, "output") {
		t.Errorf("stderr = %q, want it to name the invalid output format", stderr)
	}
}

func TestRun_LogLevelFlag_AcceptsValidLevel(t *testing.T) {
	code, _, stderr := run(t, "--log-level", "debug")

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
}

func TestRun_LogLevelFlag_RejectsInvalidLevel(t *testing.T) {
	code, stdout, stderr := run(t, "--log-level", "bogus")

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "log-level") {
		t.Errorf("stderr = %q, want it to name the invalid log-level", stderr)
	}
}

func TestRun_GlobalFlags_AllParse(t *testing.T) {
	// Every global flag must at least parse. Their effects land in later beads
	// (logging F4, config-load F6); here we only assert the flag set is complete
	// and accepted. --config names a path but is not loaded yet, so this is fine.
	code, _, stderr := run(t,
		"--debug", "--no-color", "--no-input",
		"--config", "somewhere.yaml", "-o", "json",
	)

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
}

func TestRun_QuietFlag_Parses(t *testing.T) {
	code, _, stderr := run(t, "-q")

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
}

func TestRun_DashV_IsVerboseNotVersion(t *testing.T) {
	// -v is reclaimed for --verbose, so it must NOT short-circuit to the version
	// line; with no subcommand it is accepted and the root prints help.
	code, stdout, stderr := run(t, "-v")

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("stdout = %q; -v should be verbose (help shown), not print the version", stdout)
	}
}

func TestRun_UnknownFlag_IsUsageErrorOnStderr(t *testing.T) {
	code, stdout, stderr := run(t, "--nope")

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty (errors go to stderr, ADR-0002)", stdout)
	}
	if !strings.Contains(stderr, "unknown flag") {
		t.Errorf("stderr = %q, want it to mention the unknown flag", stderr)
	}
}

func TestRun_UnknownCommand_IsUsageError(t *testing.T) {
	code, stdout, stderr := run(t, "nope")

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if stderr == "" {
		t.Error("stderr is empty, want an error message for the rejected argument")
	}
}

func TestRun_NoArgs_ShowsHelpOnStdout(t *testing.T) {
	code, stdout, stderr := run(t)

	if code != cli.ExitOK {
		t.Errorf("exit code = %d, want %d (ExitOK)", code, cli.ExitOK)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("stdout = %q, want help output (containing %q)", stdout, "Usage:")
	}
}
