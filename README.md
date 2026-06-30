# go_cli_tool_template

A scaffold for production-grade Go OSS command-line tools. It is three things in
one repository (ADR-0001):

1. **A compilable, CI-tested Template** (`template/`) — a real Go CLI module with
   the plumbing every serious tool needs (subcommands, global flags, layered
   config, structured logging, progress, a strict exit-code contract). Because it
   builds and passes tests, it cannot bit-rot.
2. **A deterministic scaffold** (`cmd/scaffold`) that instantiates the Template
   into a brand-new project — copying it, substituting placeholder tokens, and
   wiring up `go mod tidy` + `git init`.
3. **Guides and decision records** (`docs/`) explaining *why* the Template is
   shaped the way it is.

The result: you scaffold a new tool from known-good defaults in seconds, then
replace one placeholder package with your real logic.

---

## Repository layout

This is a three-module Go workspace (`go.work`):

| Path | Module | What it is |
|------|--------|------------|
| `./` | `…/go_cli_tool_template` | the repo root; hosts `cmd/scaffold` |
| `./template` | `github.com/OWNER/REPLACE_TOOL` | the Template — a self-contained CLI module with **sentinel** placeholders |
| `./examples/yaml-validator` | `github.com/NickMoignard/yamlvalidate` | a **worked example**: the Template instantiated and extended into a real JSON Schema YAML validator |

`template/` deliberately uses placeholder strings (its module path and tool name)
and is its own module, so the scaffold copies *only* it — never these meta-files.
It must build standalone, outside the workspace:

```console
$ cd template && GOWORK=off go build ./...
```

## Quickstart: scaffold a new tool

From the repo root, run the scaffold (it copies `template/` by default):

```console
$ go run ./cmd/scaffold \
    -dest   ../my-tool \
    -module github.com/you/my-tool \
    -name   my-tool \
    -author "Your Name" \
    -description "What my-tool does"
$ cd ../my-tool && go test ./...
```

You now have a compiling, tested CLI named `my-tool` with a placeholder `check`
command. Swap the body of `internal/check` (and rename the command) for your real
logic and you have a real tool. See
[docs/guides/instantiating-a-new-tool.md](docs/guides/instantiating-a-new-tool.md)
for the full flag set and how the worked example does exactly this.

Working with an agent? Invoke the **`scaffold-go-cli`** skill
([`skills/scaffold-go-cli`](skills/scaffold-go-cli/SKILL.md)) — it
collects these inputs and runs the scaffold for you. See
[Installing the scaffold-go-cli skill](#installing-the-scaffold-go-cli-skill)
below.

## Installing the scaffold-go-cli skill

The skill lives at [`skills/scaffold-go-cli/SKILL.md`](skills/scaffold-go-cli/SKILL.md).
Install it into your agent however you prefer:

**npx skills (recommended, agent-agnostic).** Auto-discovers the top-level
`skills/` dir and places the skill in the right location (`~/.claude/skills`,
etc.):

```console
$ npx skills add NickMoignard/go_cli_tool_template
```

If you need to disambiguate, scope it explicitly with
`npx skills add NickMoignard/go_cli_tool_template --skill scaffold-go-cli`.

**curl | sh (no npm required).** Installs `SKILL.md` into
`${CLAUDE_SKILLS_DIR:-$HOME/.claude/skills}/scaffold-go-cli`:

```console
$ curl -fsSL https://raw.githubusercontent.com/NickMoignard/go_cli_tool_template/master/install.sh | sh
```

From a local clone the same script copies the in-repo `SKILL.md` instead of
downloading it:

```console
$ sh install.sh
```

**Claude Code plugin.** Claude Code users can also install via
`claude plugin marketplace add NickMoignard/go_cli_tool_template`.

## What you get out of the box

The Template (and therefore every scaffolded tool) ships with:

- **A clean command tree** on [spf13/cobra](https://github.com/spf13/cobra) +
  pflag — the de-facto standard, instantly recognisable to OSS reviewers
  (ADR-0004).
- **A strict exit-code contract** — `0` valid · `1` validation failure · `2`
  usage error · `>2` internal — so scripts and CI can tell "the input is wrong"
  from "the tool broke" (ADR-0002).
- **A global flag set**: `-o/--output` (text/json), `--log-level` /
  `-v/--verbose` / `--debug` / `-q/--quiet`, `--log-format`, `--no-color`,
  `--config`, `--no-input`, `--version`.
- **Layered configuration** — `flags > env > config file > defaults` — via a
  small stdlib + yaml loader (no viper).
- **Structured logging** (`slog` + tint) and a **progress** abstraction that
  no-ops when not attached to a terminal.
- **A test net**: table tests, golden files, and end-to-end
  [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)
  scenarios, all green under `-race`.

The runtime contract is documented in
[docs/guides/cli-conventions.md](docs/guides/cli-conventions.md).

## Documentation map

- **[CONTEXT.md](CONTEXT.md)** — glossary of the project's vocabulary (Template,
  Scaffold, Sentinel tokens, Worked example, …). Start here for terminology.
- **[docs/guides/](docs/guides/)** — the "why" guides (CLI conventions,
  instantiating a new tool).
- **[docs/adr/](docs/adr/)** — Architecture Decision Records, the load-bearing
  choices:
  - [0001](docs/adr/0001-repo-is-template-skill-and-guides.md) — one repo holds
    Template + Skill + Guides
  - [0002](docs/adr/0002-validator-exit-code-contract.md) — the exit-code
    contract
  - [0003](docs/adr/0003-deterministic-go-scaffold-script.md) — a deterministic
    Go scaffold (not gonew / not the LLM)
  - [0004](docs/adr/0004-cli-framework-cobra-fang.md) — cobra + pflag, fang
    isolated, viper excluded
- **[examples/yaml-validator/](examples/yaml-validator/)** — a complete, runnable
  tool built from the Template; its README walks the workflow.

## Project status

v1 is complete. Built, tested, and documented: the Template (command tree, global
flags, the exit-code contract, layered config, `slog` logging, progress, and a
placeholder `check` command) with [fang](https://github.com/charmbracelet/fang)
styled help/errors behind a swappable seam, shell completions, and per-subcommand
man pages; the deterministic scaffold and the `scaffold-go-cli` agent skill that
wraps it; the worked YAML-validator example; repo CI (lint/test/vuln across all
modules) plus a single-module CI and a tag-driven
[GoReleaser](https://goreleaser.com) release pipeline (archives, checksums, GitHub
Releases, `go install`, Homebrew tap, cosign + SBOM) that ship with every
scaffolded project; community health files; and the guides under `docs/`.

Tracked as future enhancements (run `bd ready`): a CI drift-check for the example
and extra packaging channels (Scoop, nFPM). This project uses
**[beads](https://github.com/gastownhall/beads)** for issue tracking.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) (note: work is tracked with `bd`, not
GitHub issues) and the [Code of Conduct](CODE_OF_CONDUCT.md). Report security
issues per [SECURITY.md](SECURITY.md). The Template carries its own community
files (`template/LICENSE`, `README`, `CONTRIBUTING`, …), so scaffolded projects
are OSS-ready from the first commit.

## License

[MIT](LICENSE) © Nick Moignard
