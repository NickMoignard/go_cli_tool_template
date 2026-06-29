package scaffold_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NickMoignard/go_cli_tool_template/internal/scaffold"
)

// realTemplateDir returns the repo's template/ directory (two levels up from this
// package), skipping the test if it is not present.
func realTemplateDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "template"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		t.Skipf("real template dir not found: %v", err)
	}
	return dir
}

// TestGenerate_RealTemplateCompiles instantiates the actual template/ and checks
// the generated project builds standalone (outside the workspace), which is the
// core promise of the scaffold: substitution always yields a compiling tool.
func TestGenerate_RealTemplateCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build-based integration test in -short mode")
	}
	dest := filepath.Join(t.TempDir(), "widget")
	spec := scaffold.Spec{
		Source:      realTemplateDir(t),
		Dest:        dest,
		Module:      "github.com/alice/widget",
		Name:        "widget",
		Author:      "Alice Example",
		Year:        "2026",
		Description: "A widget validator.",
	}
	if err := scaffold.Generate(spec); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// No sentinel tokens may survive anywhere in the generated tree.
	assertNoSentinels(t, dest)

	// The renamed entrypoint must exist; the sentinel-named one must not.
	if _, err := os.Stat(filepath.Join(dest, "cmd", "widget", "main.go")); err != nil {
		t.Errorf("cmd/widget/main.go missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "cmd", "REPLACE_TOOL")); !os.IsNotExist(err) {
		t.Errorf("cmd/REPLACE_TOOL should have been renamed")
	}

	// It must build standalone, outside the go.work workspace.
	build := exec.Command("go", "build", "./...")
	build.Dir = dest
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("generated project failed to build: %v\n%s", err, out)
	}
}

func assertNoSentinels(t *testing.T, root string) {
	t.Helper()
	tokens := []string{"github.com/OWNER/REPLACE_TOOL", "REPLACE_TOOL", "REPLACE_AUTHOR", "REPLACE_YEAR", "REPLACE_DESCRIPTION"}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, tok := range tokens {
			if strings.Contains(string(b), tok) {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("sentinel %q survived in %s", tok, rel)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
