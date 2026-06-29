package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NickMoignard/go_cli_tool_template/internal/scaffold"
)

// writeTree materializes files (relative path -> content) under dir.
func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func fakeTemplate(t *testing.T) string {
	t.Helper()
	src := filepath.Join(t.TempDir(), "template")
	writeTree(t, src, map[string]string{
		"go.mod":                   "module github.com/OWNER/REPLACE_TOOL\n\ngo 1.25.0\n",
		"cmd/REPLACE_TOOL/main.go": "package main // REPLACE_TOOL\n",
		"internal/cli/cli.go":      "package cli\nimport _ \"github.com/OWNER/REPLACE_TOOL/internal/version\"\n",
	})
	return src
}

func TestGenerate_ReplacesTokensAndRenamesCmdDir(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "widget")
	spec := scaffold.Spec{
		Source: fakeTemplate(t),
		Dest:   dest,
		Module: "github.com/alice/widget",
		Name:   "widget",
	}
	if err := scaffold.Generate(spec); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if got := read(t, filepath.Join(dest, "go.mod")); got != "module github.com/alice/widget\n\ngo 1.25.0\n" {
		t.Errorf("go.mod not rewritten: %q", got)
	}
	// cmd/REPLACE_TOOL must be renamed; the old path must not exist.
	if _, err := os.Stat(filepath.Join(dest, "cmd", "REPLACE_TOOL")); !os.IsNotExist(err) {
		t.Errorf("cmd/REPLACE_TOOL still present, want renamed")
	}
	main := filepath.Join(dest, "cmd", "widget", "main.go")
	if got := read(t, main); got != "package main // widget\n" {
		t.Errorf("renamed main.go content = %q", got)
	}
	if got := read(t, filepath.Join(dest, "internal", "cli", "cli.go")); got != "package cli\nimport _ \"github.com/alice/widget/internal/version\"\n" {
		t.Errorf("import path not rewritten: %q", got)
	}
}

func TestGenerate_SkipsGitDir(t *testing.T) {
	src := fakeTemplate(t)
	writeTree(t, src, map[string]string{".git/config": "[core]\n"})
	dest := filepath.Join(t.TempDir(), "widget")
	spec := scaffold.Spec{Source: src, Dest: dest, Module: "github.com/alice/widget", Name: "widget"}
	if err := scaffold.Generate(spec); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git copied into generated project, want skipped")
	}
}

func TestGenerate_RefusesNonEmptyDest(t *testing.T) {
	dest := t.TempDir() // exists and... make it non-empty
	writeTree(t, dest, map[string]string{"existing.txt": "x"})
	spec := scaffold.Spec{Source: fakeTemplate(t), Dest: dest, Module: "github.com/alice/widget", Name: "widget"}
	if err := scaffold.Generate(spec); err == nil {
		t.Fatal("Generate into non-empty dest: want error, got nil")
	}
}
