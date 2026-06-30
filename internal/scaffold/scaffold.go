// Package scaffold instantiates the ./template module into a brand-new Go CLI
// project. It copies the template tree, replaces the sentinel tokens (ADR-0003),
// renames the cmd/<tool> directory, and (optionally) runs `go mod tidy` and
// `git init`. Everything here is deterministic so the output can be golden-tested.
package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Spec describes one instantiation of the Template.
type Spec struct {
	Source      string // path to the template/ directory to copy from
	Dest        string // path to the new project directory to create
	Module      string // new module path, e.g. github.com/alice/widget
	Name        string // tool/binary name, e.g. widget
	Author      string // copyright holder for community files
	Year        string // copyright year
	Description string // one-line tool description
	Email       string // contact/maintainer email for community + release files
}

// Generate copies the template tree at s.Source into s.Dest, substituting sentinel
// tokens in both file contents and path segments. The destination must not already
// contain files. A .git directory in the source is never copied. File permissions
// are preserved.
func Generate(s Spec) error {
	if err := ensureEmptyDir(s.Dest); err != nil {
		return err
	}
	repls := s.replacements()

	return filepath.WalkDir(s.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(s.Source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		// Never copy version-control metadata into the generated project.
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		dst := filepath.Join(s.Dest, relDest(rel, repls))
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, []byte(apply(string(content), repls)), info.Mode().Perm())
	})
}

// ensureEmptyDir creates dest if absent and verifies it holds no entries.
func ensureEmptyDir(dest string) error {
	if dest == "" {
		return fmt.Errorf("scaffold: empty destination path")
	}
	entries, err := os.ReadDir(dest)
	if os.IsNotExist(err) {
		return os.MkdirAll(dest, 0o755)
	}
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("scaffold: destination %q is not empty", dest)
	}
	return nil
}
