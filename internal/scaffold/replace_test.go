package scaffold

import "testing"

// The module-path sentinel embeds the tool-name sentinel
// (github.com/OWNER/REPLACE_TOOL contains REPLACE_TOOL), so the replacements
// must be applied longest-first or the module path gets mangled.
func TestSpecReplacements_Ordering(t *testing.T) {
	s := Spec{Module: "github.com/alice/widget", Name: "widget"}
	repls := s.replacements()

	in := "import \"github.com/OWNER/REPLACE_TOOL/internal/cli\" // run REPLACE_TOOL\n"
	want := "import \"github.com/alice/widget/internal/cli\" // run widget\n"
	if got := apply(in, repls); got != want {
		t.Errorf("apply()\n got = %q\nwant = %q", got, want)
	}
}

func TestSpecReplacements_AllSentinels(t *testing.T) {
	s := Spec{
		Module:      "github.com/alice/widget",
		Name:        "widget",
		Author:      "Alice Example",
		Year:        "2026",
		Description: "A widget that widgets.",
	}
	cases := []struct{ in, want string }{
		{sentinelModule, "github.com/alice/widget"},
		{sentinelName, "widget"},
		{sentinelAuthor, "Alice Example"},
		{sentinelYear, "2026"},
		{sentinelDescription, "A widget that widgets."},
	}
	repls := s.replacements()
	for _, c := range cases {
		if got := apply(c.in, repls); got != c.want {
			t.Errorf("apply(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRelDest_RenamesPathTokens(t *testing.T) {
	s := Spec{Module: "github.com/alice/widget", Name: "widget"}
	repls := s.replacements()
	cases := []struct{ in, want string }{
		{"cmd/REPLACE_TOOL/main.go", "cmd/widget/main.go"},
		{"internal/cli/cli.go", "internal/cli/cli.go"},
		{"go.mod", "go.mod"},
	}
	for _, c := range cases {
		if got := relDest(c.in, repls); got != c.want {
			t.Errorf("relDest(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
