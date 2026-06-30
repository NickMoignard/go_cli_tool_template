package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// The fang seam is swappable for stock cobra (ADR-0004). Swapping in plainExecute
// must still run the command and render its error to stderr — proving fang is not
// load-bearing and the exit-code mapping is identical either way.
func TestExecuteSeam_PlainExecuteSwap(t *testing.T) {
	orig := execute
	execute = plainExecute
	t.Cleanup(func() { execute = orig })

	var stdout, stderr bytes.Buffer
	code := runCmd(context.Background(), NewRootCmd(), []string{"--nope"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("stderr = %q, want plain cobra rendering (containing %q)", stderr.String(), "error:")
	}
	if stdout.String() != "" {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}
