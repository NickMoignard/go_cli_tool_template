package scaffold

import "strings"

// Sentinel tokens (ADR-0003): collision-free placeholder strings seeded into the
// Template that the scaffold find-and-replaces during instantiation. This block is
// the single source of truth for the token contract — files in template/ (and the
// community files) must use exactly these strings, and the scaffold substitutes
// them here. Module path and tool name drive the code; author, year, description,
// and email feed the community/release files (LICENSE, README, CODE_OF_CONDUCT,
// SECURITY, the nFPM maintainer), so they are harmless no-ops where unused.
const (
	sentinelModule      = "github.com/OWNER/REPLACE_TOOL"
	sentinelName        = "REPLACE_TOOL"
	sentinelAuthor      = "REPLACE_AUTHOR"
	sentinelYear        = "REPLACE_YEAR"
	sentinelDescription = "REPLACE_DESCRIPTION"
	sentinelEmail       = "REPLACE_EMAIL"
)

// replacement is a single literal find/replace pair.
type replacement struct {
	old, new string
}

// replacements returns the ordered substitutions for this Spec. Order is
// significant: the module-path sentinel embeds the tool-name sentinel, so it must
// come first or "github.com/OWNER/REPLACE_TOOL" would be partly rewritten by the
// bare-name rule and never match.
func (s Spec) replacements() []replacement {
	return []replacement{
		{sentinelModule, s.Module},
		{sentinelName, s.Name},
		{sentinelAuthor, s.Author},
		{sentinelYear, s.Year},
		{sentinelDescription, s.Description},
		{sentinelEmail, s.Email},
	}
}

// apply rewrites content by running each replacement in order.
func apply(content string, repls []replacement) string {
	for _, r := range repls {
		content = strings.ReplaceAll(content, r.old, r.new)
	}
	return content
}

// relDest maps a template-relative path to its destination path, substituting any
// sentinel tokens that appear in path segments (e.g. cmd/REPLACE_TOOL).
func relDest(rel string, repls []replacement) string {
	return apply(rel, repls)
}
