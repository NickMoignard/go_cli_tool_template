# Guide: instantiating a new tool

This guide explains how a brand-new CLI project is produced from the Template:
what the scaffold does, the placeholder ("sentinel") tokens it substitutes, and
how the worked example demonstrates the intended workflow end to end.

For terminology see [CONTEXT.md](../../CONTEXT.md); for the design rationale see
[ADR-0003](../adr/0003-deterministic-go-scaffold-script.md).

## Why a deterministic script (not gonew, not the LLM)

Instantiation is the most mechanical step in the whole process — copy a tree,
replace some strings, rename a directory. The Template hands this to a single,
self-contained Go program (`cmd/scaffold`) rather than to `gonew` or to an LLM:

- **`gonew`** only rewrites the module path (you would still need a script for the
  rename and the author/description fields), resolves templates through the module
  proxy (it expects a *published* module, not a local `template/`), and is flagged
  experimental.
- **An LLM** doing the rename reintroduces exactly the non-determinism and drift
  that a template is meant to eliminate.

Because the scaffold *owns* the placeholder tokens, a plain string replacement is
provably correct without AST-aware tooling — and the output always compiles. The
program is itself testable; an integration test instantiates the real `template/`
and asserts the result builds standalone with zero placeholders left behind.

## Sentinel tokens

The Template is seeded with collision-free placeholder strings that the scaffold
finds and replaces. The token set is the single source of truth (defined in
`internal/scaffold/replace.go`):

| Sentinel | Replaced with | Status |
|----------|---------------|--------|
| `github.com/OWNER/REPLACE_TOOL` | your module path (`-module`) | live |
| `REPLACE_TOOL` | your tool/binary name (`-name`) | live |
| `REPLACE_AUTHOR` | copyright holder (`-author`) | reserved |
| `REPLACE_YEAR` | copyright year (`-year`, defaults to current) | reserved |
| `REPLACE_DESCRIPTION` | one-line description (`-description`) | reserved |

The module-path sentinel **embeds** the tool-name sentinel, so substitutions are
applied longest-first — the module path is rewritten before the bare name, or it
would be mangled.

The *reserved* tokens (author/year/description) are accepted by the scaffold today
but appear in no Template file yet; they become live when the community files
(LICENSE/README) that carry them are added. Replacing a token that is absent is a
harmless no-op, so the scaffold's input contract is already complete.

## Running the scaffold

From the repo root:

```console
$ go run ./cmd/scaffold \
    -dest        ../my-tool \
    -module      github.com/you/my-tool \
    -name        my-tool \
    -author      "Your Name" \
    -description "What my-tool does"
```

Flags:

| Flag | Required | Default | Meaning |
|------|----------|---------|---------|
| `-dest` | ✅ | — | destination directory (must be empty or absent) |
| `-module` | ✅ | — | new module path |
| `-name` | ✅ | — | tool/binary name |
| `-source` | | `template` | the Template directory to copy from |
| `-author` | | — | copyright holder |
| `-year` | | current year | copyright year |
| `-description` | | — | one-line description |
| `-skip-tidy` | | off | do not run `go mod tidy` in the new project |
| `-skip-git` | | off | do not run `git init` in the new project |

What it does, in order:

1. Refuse a non-empty destination, then copy the `template/` tree (skipping any
   `.git`), substituting sentinel tokens in **both file contents and path
   segments** — so `cmd/REPLACE_TOOL/` becomes `cmd/my-tool/`. File permissions
   are preserved.
2. Run `go mod tidy` in the new project (unless `-skip-tidy`).
3. Run `git init` (unless `-skip-git`).

The exit code follows the same convention as the tools it generates: `0` success,
`2` usage error (missing/!empty flags), `1` if generation, tidy, or git fails.

## From scaffold output to a real tool

A freshly scaffolded project has a placeholder `check` command in
`internal/check` and `internal/cli/check.go`. It is deliberately shaped like real
validation — iterate inputs → per-item pass/fail → exit code — but its rule is
trivial (non-empty, valid UTF-8). To build a real tool you:

1. Replace the body of the domain package (`internal/check`) with your logic,
   keeping the read-input → return-verdict shape.
2. Rename/replace the command in `internal/cli/` and wire it into the root in
   `cli.go`.
3. Update the tests, the `testscript` scenarios, and regenerate the help golden
   file (`go test ./internal/cli/ -run TestRun_Help -update`).

## The worked example

[`examples/yaml-validator/`](../../examples/yaml-validator/) is precisely this
workflow carried out: the Template instantiated (module
`github.com/NickMoignard/yamlvalidate`, name `yamlvalidate`) and then extended
into a real JSON Schema YAML validator. Its only domain change from raw scaffold
output is swapping the `check` placeholder for a `validate` command backed by
`internal/validate` (using `santhosh-tekuri/jsonschema`); everything else — flags,
config, logging, progress, exit codes — is the scaffold output untouched. It
builds and is tested on its own, proving the scaffold produces a releasable
project. Read its [README](../../examples/yaml-validator/README.md) for sample
runs.

## See also

- [CLI conventions & the runtime contract](cli-conventions.md) — what the
  generated tool does at runtime.
- [ADR-0003](../adr/0003-deterministic-go-scaffold-script.md) — the decision to
  use a deterministic script.
- [ADR-0001](../adr/0001-repo-is-template-skill-and-guides.md) — why the Template,
  Skill, and Guides live in one repo.
