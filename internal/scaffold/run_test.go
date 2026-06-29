package scaffold_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NickMoignard/go_cli_tool_template/internal/scaffold"
)

func runScaffold(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = scaffold.Run(args, &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func TestRun_MissingRequiredFlags_UsageError(t *testing.T) {
	code, _, stderr := runScaffold(t, "-source", fakeTemplate(t))
	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "module") || !strings.Contains(stderr, "name") {
		t.Errorf("stderr should name the missing flags, got: %q", stderr)
	}
}

func TestRun_HappyPath_GeneratesProject(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "widget")
	code, stdout, stderr := runScaffold(t,
		"-source", fakeTemplate(t),
		"-dest", dest,
		"-module", "github.com/alice/widget",
		"-name", "widget",
		"-skip-tidy", "-skip-git",
	)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr)
	}
	if got := read(t, filepath.Join(dest, "cmd", "widget", "main.go")); !strings.Contains(got, "widget") {
		t.Errorf("generated main.go missing tool name: %q", got)
	}
	if !strings.Contains(stdout, dest) {
		t.Errorf("stdout should report the destination, got: %q", stdout)
	}
}

func TestRun_NonEmptyDest_RuntimeError(t *testing.T) {
	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(dest, "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runScaffold(t,
		"-source", fakeTemplate(t),
		"-dest", dest,
		"-module", "github.com/alice/widget",
		"-name", "widget",
		"-skip-tidy", "-skip-git",
	)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "not empty") {
		t.Errorf("stderr should explain the failure, got: %q", stderr)
	}
}

func TestRun_DefaultsYearToCurrent(t *testing.T) {
	// With -year omitted, Run must still succeed (year defaulted internally).
	dest := filepath.Join(t.TempDir(), "widget")
	code, _, stderr := runScaffold(t,
		"-source", fakeTemplate(t),
		"-dest", dest,
		"-module", "github.com/alice/widget",
		"-name", "widget",
		"-skip-tidy", "-skip-git",
	)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr)
	}
}
