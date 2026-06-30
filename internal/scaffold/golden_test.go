package scaffold_test

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/NickMoignard/go_cli_tool_template/internal/scaffold"
)

var update = flag.Bool("update", false, "update golden files")

// goldenSpec is the fixed input the golden tests instantiate. Generate is pure
// (copy + ordered string-replace), so the same template + this spec always yield
// byte-identical output — golden-testable. Tidy/git are excluded (they are not
// part of Generate).
func goldenSpec(t *testing.T, dest string) scaffold.Spec {
	return scaffold.Spec{
		Source:      realTemplateDir(t),
		Dest:        dest,
		Module:      "github.com/example/widget",
		Name:        "widget",
		Author:      "Example Author",
		Year:        "2026",
		Description: "An example CLI.",
		Email:       "maintainer@example.com",
	}
}

// TestGenerate_Golden_Manifest locks the shape of the scaffold output — every
// output path for the fixed spec — so an added, dropped, or misrenamed file
// (e.g. cmd/widget) is caught. Regenerate intentionally with -update.
func TestGenerate_Golden_Manifest(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "widget")
	if err := scaffold.Generate(goldenSpec(t, dest)); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	var paths []string
	err := filepath.WalkDir(dest, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, relErr := filepath.Rel(dest, p)
			if relErr != nil {
				return relErr
			}
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	assertGolden(t, "manifest.golden", strings.Join(paths, "\n")+"\n")
}

// TestGenerate_Golden_KeyFiles locks the exact substituted content of the files
// where token replacement is load-bearing — together they exercise all six
// sentinels (module, name, author, year, description, email — the last via
// SECURITY.md's maintainer contact).
func TestGenerate_Golden_KeyFiles(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "widget")
	if err := scaffold.Generate(goldenSpec(t, dest)); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, rel := range []string{"go.mod", "cmd/widget/main.go", "LICENSE", "README.md", "SECURITY.md"} {
		golden := "keyfiles/" + strings.ReplaceAll(rel, "/", "_") + ".golden"
		assertGolden(t, golden, read(t, filepath.Join(dest, filepath.FromSlash(rel))))
	}
}

// assertGolden compares actual against testdata/<name>, or rewrites it under
// -update.
func assertGolden(t *testing.T, name, actual string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(actual), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s: %v (run `go test ./internal/scaffold/ -update`)", path, err)
	}
	if actual != string(want) {
		t.Errorf("scaffold output for %s differs from golden; run -update to inspect the diff", name)
	}
}
