# Instantiation uses one self-contained Go scaffold script, not gonew or the LLM

The Template is instantiated by a single, self-contained Go program (e.g. `cmd/scaffold`)
that does *all* mechanical work deterministically: copy `template/`, find-and-replace a set
of owned **Sentinel tokens** (module path, tool/binary name, author, year, description),
rename the `cmd/<tool>` directory, then run `go mod tidy` and `git init`. Because we own the
sentinel tokens, a plain string-replace is correct without AST-aware tooling, and the result
always compiles. The Skill only collects human inputs and invokes the script.

## Considered Options

- **`gonew`** (official Go template tool) — rejected as the mechanism: it rewrites *only*
  the module path (still needs a script for the rename/author/etc.), resolves templates
  through the module proxy (expects a published module, not a local `template/`), and is
  flagged experimental. Relegated to a Guide footnote for the module-path-only case.
- **LLM-driven rename/substitution** — rejected: reintroduces the non-determinism and
  drift that choosing a script exists to prevent. Renaming is the most mechanical step and
  the last thing to hand to a fuzzy executor.

## Consequences

The scaffold program is itself testable (golden-file its output). The hand-maintained
`examples/yaml-validator/` can drift from `template/`; a future CI check should re-run the
scaffold and diff the non-domain files.
