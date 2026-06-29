---
status: accepted
supersedes: 0004-cli-framework-deferred (placeholder)
---

# CLI framework: spf13/cobra + pflag, with charmbracelet/fang as an isolated add-on, no viper

The Template builds on **spf13/cobra** (+ **pflag**) as its base CLI framework. **charmbracelet/fang**
is adopted as an *optional, isolated* styling layer over cobra, and **spf13/viper is excluded** in favour
of a small stdlib config loader (per ADR earlier in the I/O/config decisions). This replaces the
placeholder that deferred the choice to research; it is the output of a multi-source, adversarially-verified
deep-research pass (mid-2026).

## Why cobra

Cobra is the only candidate that satisfies *all* the template's hard requirements natively and in one place:

- **Subcommands + flags:** native nested subcommands and fully POSIX/GNU-style short+long flags via pflag —
  exactly the global flag set we specified (`--log-level`/`--verbose`/`--debug`, `-o/--output`, `-q/--quiet`,
  `--no-color`, `--config`, `--no-input`, `--version`, `-h`).
- **Shell completions (a hard requirement):** built into the binary for **all four** required shells —
  bash (incl. V2), zsh, fish, powershell — via the portable `ValidArgsFunction`/`RegisterFlagCompletionFunc`
  path. No add-on needed.
- **Man pages (a hard requirement):** `cobra/doc.GenManTree` generates **one man page per (sub)command**
  from the full command tree (via go-md2man, pulled in only by the optional `cobra/doc` subpackage).
- **Recognizability (the audience constraint):** the de-facto standard — Kubernetes, Hugo, GitHub CLI — so
  newcomers learn *the* idiomatic structure and OSS reviewers recognise it instantly.
- **I/O control:** `SetOut`/`SetErr`, `SilenceErrors`/`SilenceUsage`, and `RunE` returning errors let the thin
  `cmd/` entrypoint own the stdout-vs-stderr split and the custom exit-code mapping (0/1/2/>2) cleanly.

Latest stable at research time: **cobra v1.10.2** (2025-12-04), actively maintained (v1.10.3 in pre-release
April 2026).

## Why fang — but isolated and swappable, not load-bearing

Fang wraps a `*cobra.Command` (`fang.Execute(...)`) to fix cobra's one real weakness — ugly default help — with
fully styled help/usage/error output and silent-usage-after-error. It also bundles `completion` and `man`
commands. Latest stable **v2.0.1** (2026-03-11).

The catch: fang is **self-described as experimental** and jumped v0→v1→v2 within days in March 2026, so its API
is a real churn risk for a long-lived template. Therefore:

- Fang is **opt-in and isolated behind a thin seam** (a single styling/execute wrapper), so it can be swapped
  for cobra's plain `Execute()` without touching command definitions. It is *not* load-bearing.
- Its version is **pinned**.
- **Man pages use cobra's native `doc.GenManTree`, not fang's man command.** Fang's man generation (via
  `muesli/mango`) emits a *single whole-tool* page; our requirement is *per-subcommand* pages from the command
  tree, which only `doc.GenManTree` delivers. Fang is used for styled help + (optionally) completions only.

## Why not viper

Viper's precedence *does* match our intent (flags > env > file > defaults), but it is a heavyweight "complete
12-Factor configuration solution" (remote stores, hot-reload, 5+ formats) — all unused for a
flags>env>YAML>defaults loader — and it carries a known precedence gotcha where a defaulted pflag overrides an
env var (spf13/viper#671). A small stdlib loader fits the lean-dependency posture. **koanf** is noted as the
lean, modular alternative *if* a config library is ever wanted (deps detached per provider/parser); the
threshold for reaching for it would be a genuine need for remote config or hot-reload.

## Comparison (mid-2026)

| | cobra (+pflag) | urfave/cli v3 | alecthomas/kong | stdlib flag |
|---|---|---|---|---|
| Nested subcommands | ✅ native | ✅ native | ✅ native | ❌ hand-rolled |
| Completions (4 shells) | ✅ built-in, static | ✅ dynamic (runtime) | partial | ❌ |
| Man-page gen (command tree) | ✅ `cobra/doc` (per-cmd) | ⚠️ separate `cli-docs` module | ✅ via `mango-kong` | ✅ via `mango` |
| Dependency weight | moderate | near-zero runtime deps | low | none |
| Maturity / adoption | ✅ de-facto standard | mature, less ubiquitous | niche | stdlib |
| Help customization | weak by default → **fang** fixes it | moderate | moderate | manual |
| Learning curve | moderate (lots of concepts) | low | low (struct tags) | trivial |

## Risks (recommended stack)

- **Fang experimental / API churn** — mitigated by isolation + version pin (above).
- **cobra man-page edge case** — hyphenated command names can yield ambiguous `.1` filenames; avoid or post-process.
- **PowerShell completion** requires PowerShell ≥ 5.0 (a release-note "7.2+" claim was found imprecise).
- **Licenses:** cobra = Apache-2.0, pflag = BSD-3-Clause, fang = MIT — all permissive and compatible with the
  Template's MIT default. (Re-confirm at adoption.)

## Versions to pin at scaffold time

cobra **v1.10.2**, fang **v2.0.1**, (pflag tracks cobra). Re-verify latest stables when the Template is built —
these are pinned to the mid-2026 research date.
