package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OWNER/REPLACE_TOOL/internal/cli"
)

func TestRun_Man_GeneratesPerCommandPages(t *testing.T) {
	// A not-yet-existing dir: the command must create it.
	dir := filepath.Join(t.TempDir(), "manpages")

	code, _, stderr := run(t, "man", dir)

	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
	// One .1 per command: the root and each real subcommand.
	for _, name := range []string{"REPLACE_TOOL.1", "REPLACE_TOOL-check.1"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected man page %s: %v", name, err)
		}
	}
}

func TestRun_Completion_AllShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			code, stdout, stderr := run(t, "completion", shell)
			if code != cli.ExitOK {
				t.Fatalf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
			}
			if stdout == "" {
				t.Errorf("completion %s produced no script on stdout", shell)
			}
		})
	}
}
