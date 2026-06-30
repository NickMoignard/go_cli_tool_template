package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// newManCmd returns a hidden command that writes one man page per (sub)command
// into <dir> via cobra/doc.GenManTree (go-md2man). Man pages come from cobra/doc,
// not fang, because only GenManTree emits a page *per command* from the tree
// (ADR-0004). It is hidden because it is a build/release helper rather than a
// day-to-day command — GoReleaser runs it to package the .1 pages.
//
// Keep subcommand names free of '-': GenManTree joins the command path with '-'
// to name each .1 file, so a hyphenated subcommand would yield an ambiguous
// filename (a documented GenManTree limitation).
func newManCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "man <dir>",
		Short:  "Generate man pages into <dir>",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(args[0], 0o755); err != nil {
				return internalErr(err)
			}
			header := &doc.GenManHeader{Section: "1"}
			if err := doc.GenManTree(cmd.Root(), header, args[0]); err != nil {
				return internalErr(err)
			}
			return nil
		},
	}
}
