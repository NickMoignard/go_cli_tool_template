# Context: go_cli_tool_template

Glossary of terms for this project. Definitions only — no implementation details.

## Glossary

### Template
The canonical, CI-tested, compilable Go CLI project living in this repo. It is real
code that builds and passes tests, so it cannot bit-rot. It is the source of truth that
the Skill instantiates from.

### Skill
The Claude Code agent skill (`SKILL.md` + supporting files) that instantiates the
Template into a brand-new Go CLI project — performing renames, substitutions, and
running setup commands. Thin orchestration over known-good sources.

### Runbook
The procedural, step-by-step content (primarily the SKILL.md body) describing how to
scaffold and stand up a new CLI tool. The "do these steps in this order" layer.

### Guides
Reference documentation in `docs/` explaining the *why* behind the best practices the
Template encodes (distribution, CI/CD, I/O conventions, flags). Background reading,
distinct from the Runbook's procedure.

### Deliverable shape
This repo is simultaneously (a) the compilable reference Template, (b) the home of the
Skill that instantiates it, and (c) a `docs/` set of Guides. All three ship together.

### Scaffold script
The single, self-contained Go program (lives at the repo root, e.g. `cmd/scaffold`) that
performs ALL mechanical instantiation work deterministically: copy `template/`, replace
the Sentinel tokens, rename the `cmd/<tool>` directory, run `go mod tidy` and `git init`.
The Skill only collects human inputs and invokes it; the script — not the LLM — owns
correctness. Distinct from the Template, which is the code being instantiated.

### Placeholder domain
The trivial command that the generic Template ships with. It exercises the full I/O
contract (streams, `--output`, progress, exit codes 0/1/2, config, logging) but with
trivial logic and no heavy dependencies. Deliberately shaped to mirror a `validate`
command (iterate files → per-item pass/fail → exit code) so swapping in real domain
logic is obvious. Keeps the Template generic — instantiable into any CLI, not just a
validator.

### Worked example
The fleshed-out reference tool at `examples/yaml-validator/` — the Template instantiated
and extended into a real JSON Schema YAML validator, built and tested by CI. Proves the
Scaffold script produces a releasable tool and demonstrates the intended workflow.
Distinct from the Placeholder domain (trivial, generic) and from the Template itself.

### Sentinel tokens
The collision-free placeholder strings seeded into the Template (module path, tool/binary
name, author, year, description) that the Scaffold script find-and-replaces during
instantiation. Owning these tokens is what lets a plain string-replace be correct without
AST-aware tooling like `gonew`.
