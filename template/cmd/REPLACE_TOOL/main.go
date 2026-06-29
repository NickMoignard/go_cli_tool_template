// Command REPLACE_TOOL is the generated CLI's entrypoint.
//
// REPLACE_TOOL is a SENTINEL token (ADR-0003): the scaffold renames this
// directory and substitutes the real tool name throughout. Per ADR-0002 the
// entrypoint stays thin — it only wires the command tree and maps results to
// exit codes (0 valid / 1 invalid / 2 usage / >2 internal).
//
// This is a build placeholder. The cobra skeleton + internal/version land in
// beads go_cli_tool_template-804 (F2); the I/O contract in -ceg (F3).
package main

func main() {}
