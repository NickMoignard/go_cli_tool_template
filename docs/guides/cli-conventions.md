# Guide: CLI conventions & the runtime contract

This guide explains the runtime behaviour every tool scaffolded from the Template
inherits: how it uses the streams, what its exit codes mean, the global flags,
and how configuration, logging, and progress fit together. It is background
reading — the *why* behind the code in `template/internal/`.

For the terminology used here (Template, Placeholder domain, …) see
[CONTEXT.md](../../CONTEXT.md).

## Streams: stdout is data, stderr is everything else

The Template keeps a strict separation:

- **stdout** carries the tool's *primary output* — the machine-consumable result
  (the report, the JSON, the thing you would pipe into another command). Nothing
  else is written there.
- **stderr** carries logs, progress bars, and error messages — the human-facing
  chatter that must never corrupt a pipe.

This is why the command layer threads `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`
everywhere instead of calling `fmt.Println`: it keeps the split testable and lets
`tool ... > result.json` Just Work.

## Exit codes (ADR-0002)

Exit codes are a public, scripted contract — painful to change once users depend
on them — so the Template fixes them up front:

| Code | Meaning | Example |
|------|---------|---------|
| `0`  | success / valid | every input conformed |
| `1`  | **validation failure** — ran fine, input did not conform | a file failed its schema |
| `2`  | usage error — bad flags, missing/unreadable file, unknown command | `--nope`, a missing input |
| `>2` | unexpected / internal error | a bug or I/O failure (`3` = internal) |

The non-obvious part: **a validation failure is not a tool error.** It gets its
own code (`1`), distinct from both success and a crash, so a script can do
`tool validate config.yaml && deploy` and CI can tell "the file is wrong" apart
from "the tool broke".

How it is implemented (see `template/internal/cli/cli.go`):

- Domain code never calls `os.Exit`. It returns typed errors that satisfy a
  `CodedError` interface (`usageError` → 2, `validationError` → 1,
  `internalError` → 3).
- The thin `cmd/` entrypoint is a one-liner around `cli.Run(...) int`; `Run` maps
  the returned error to a code in exactly one place.
- An *uncoded* error reaching `Run` came from cobra's flag/arg parsing, so it is
  treated as a usage error (`2`).

Because the mapping lives in one function, it is unit-tested directly *and*
pinned end-to-end with `testscript` scenarios under
`internal/cli/testdata/script/`.

## Global flags

These persistent flags are available on every (sub)command:

| Flag | Default | Purpose |
|------|---------|---------|
| `-o, --output` | `text` | output format: `text` or `json` |
| `--log-level` | `warn` | base verbosity: `error`/`warn`/`info`/`debug` |
| `-v, --verbose` | off | raise level to at least `info` |
| `--debug` | off | raise level to `debug` |
| `-q, --quiet` | off | force level to `error` and suppress progress |
| `--log-format` | `auto` | `auto` (TTY-aware) / `text` / `json` |
| `--no-color` | off | disable ANSI colour (also honours `NO_COLOR` and non-TTY) |
| `--config` | — | explicit config file path |
| `--no-input` | off | never prompt; assume non-interactive |
| `--version` | — | print version and exit |

`--verbose`, `--debug`, and `--quiet` are aliases that resolve into a single
effective log level: start at `--log-level`, raise it for verbose/debug (the most
verbose request wins), then `--quiet` overrides everything to `error`. Invalid
flag *values* (e.g. `--output xml`) are usage errors (exit `2`), validated once
in the root `PersistentPreRunE` so the rule holds for every subcommand.

## Configuration precedence

Configuration resolves with the precedence:

```
flags > environment > config file > defaults
```

- **Defaults** are the built-in values (the same as the flag defaults above).
- **Config file** is YAML. When `--config` is given it is the only file
  consulted; otherwise the tool searches the current directory then the user
  config dir (`$XDG_CONFIG_HOME/<app>/`) for `.<app>.yaml` then `config.yaml`,
  first match wins. Keys: `output`, `log_level`, `log_format`, `no_color`.
- **Environment** overrides the file. Variables are prefixed with the uppercased
  tool name: `<APP>_OUTPUT`, `<APP>_LOG_LEVEL`, `<APP>_LOG_FORMAT`,
  `<APP>_NO_COLOR`.
- **Flags** win over everything — but *only when explicitly set*. The config
  layer is overlaid first, then flags the user actually passed are re-applied on
  top, so an unset flag does not clobber a config-file value with its default.

The loader is a small stdlib + `yaml.v3` package (`internal/config`), not viper —
a deliberate lean-dependency choice (ADR-0004). It is fully injectable
(`Source{Environ, SearchDirs, …}`) so it is tested without touching the real
environment or filesystem.

## Logging

Logging uses the standard library's `slog` with [tint](https://github.com/lmittmann/tint)
for readable, optionally-coloured console output. The format follows
`--log-format`: `auto` picks tinted text on a TTY and JSON otherwise; `text` and
`json` force the choice. The effective level comes from the resolved flag set
(above). The logger is built once in the root pre-run and passed to subcommands
via the command context — logs go to **stderr**, never stdout.

## Progress

Long-running work reports through a small `Progress` interface, so the call site
does not care whether a bar is actually drawn. The concrete implementation wraps
[progressbar](https://github.com/schollz/progressbar); when output is not a
terminal, or under `--quiet`, it resolves to a **no-op** so piped/CI output stays
clean. Like logs, progress is drawn on stderr.

## See also

- [Instantiating a new tool](instantiating-a-new-tool.md) — how the scaffold turns
  the Template into a real project.
- [ADR-0002](../adr/0002-validator-exit-code-contract.md) — the exit-code contract.
- [ADR-0004](../adr/0004-cli-framework-cobra-fang.md) — why cobra + pflag, and why
  not viper.
