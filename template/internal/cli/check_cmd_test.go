package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OWNER/REPLACE_TOOL/internal/cli"
)

func tempFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// runIn is like run but feeds stdin, for exercising piped input.
func runIn(t *testing.T, stdin string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = cli.Run(context.Background(), args, strings.NewReader(stdin), &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func TestCheckCmd_ValidFile_ExitOK(t *testing.T) {
	f := tempFile(t, "hello, world\n")

	code, stdout, stderr := run(t, "check", f)

	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
	if !strings.Contains(strings.ToLower(stdout), "ok") {
		t.Errorf("stdout = %q, want an OK report", stdout)
	}
}

func TestCheckCmd_EmptyFile_ExitInvalid(t *testing.T) {
	f := tempFile(t, "")

	code, stdout, _ := run(t, "check", f)

	if code != cli.ExitInvalid {
		t.Errorf("exit code = %d, want %d (ExitInvalid)", code, cli.ExitInvalid)
	}
	if !strings.Contains(stdout, "fail") {
		t.Errorf("stdout = %q, want a fail report", stdout)
	}
}

func TestCheckCmd_StdinPipe_NoArgs(t *testing.T) {
	code, stdout, stderr := runIn(t, "piped content\n", "check")

	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d; stderr = %q", code, cli.ExitOK, stderr)
	}
	if !strings.Contains(stdout, "<stdin>") {
		t.Errorf("stdout = %q, want it to name <stdin>", stdout)
	}
}

func TestCheckCmd_JSONOutput(t *testing.T) {
	f := tempFile(t, "hi\n")

	code, stdout, _ := run(t, "-o", "json", "check", f)

	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d", code, cli.ExitOK)
	}
	var results []map[string]any
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if len(results) != 1 || results[0]["ok"] != true {
		t.Errorf("results = %v, want one passing result", results)
	}
}

func TestCheckCmd_MissingFile_ExitUsage(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.txt")

	code, stdout, stderr := run(t, "check", missing)

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "Cannot open") {
		t.Errorf("stderr = %q, want it to report the open failure", stderr)
	}
}

func TestCheckCmd_MixedResults_ExitInvalid(t *testing.T) {
	good := tempFile(t, "fine\n")
	bad := tempFile(t, "")

	code, stdout, _ := run(t, "check", good, bad)

	if code != cli.ExitInvalid {
		t.Errorf("exit code = %d, want %d (ExitInvalid)", code, cli.ExitInvalid)
	}
	if !strings.Contains(stdout, "ok") || !strings.Contains(stdout, "fail") {
		t.Errorf("stdout = %q, want both an ok and a fail line", stdout)
	}
}
