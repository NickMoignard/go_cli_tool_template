# Guide: the test suite

Every tool scaffolded from the Template inherits a test suite that demonstrates
three complementary styles, all in **pure stdlib `testing`** — no testify, no
assertion DSL. This guide is a map of those patterns so you can extend them as you
build your tool.

## The two seams

The CLI is written so behaviour is testable without a subprocess:

- **`cli.Run(ctx, args, stdin, stdout, stderr) int`** — the black-box seam. Tests
  drive the whole program through it with `bytes.Buffer`s and assert on the exit
  code and stream contents, exactly as a user (or script) sees it.
- **`runCmd(ctx, root, …)`** — a white-box seam (package `cli`) that runs a
  *prepared* root, so a test can attach an extra subcommand and exercise the
  error→exit-code mapping directly.

`main()` is a one-liner around `Run`, so there is nothing left in it to test.

## 1. Table-driven tests

The default for pure logic: a slice of cases, one `t.Run` per case. See
`internal/cli/options_test.go` (log-level resolution), `exitcode_test.go` (the
typed-error → exit-code mapping), and `internal/config` (precedence). Reach for
this whenever you can enumerate inputs and expected outputs.

## 2. Golden-file tests

For rich, multi-line output that is tedious to inline, assert against a file under
`testdata/` and regenerate it with a `-update` flag. The Template golden-tests the
root `--help` output (`internal/cli/cli_test.go` → `testdata/help.golden`):

```console
$ go test ./internal/cli/ -run TestRun_Help -update   # rewrite the golden
```

Because help is fang-styled, that test pins the width (fang's `__FANG_TEST_WIDTH`)
so the golden is deterministic across terminals. Review a golden diff like any
other code change — an unexpected diff is a regression.

## 3. Testscript (.txtar) end-to-end

`internal/cli/testdata/script/*.txtar` exercise the built command end-to-end via
[rogpeppe/go-internal/testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript):
real streams, real process semantics. `TestMain` registers the binary in-process,
so these run without compiling a separate executable. Each script asserts
success/failure and stream contents — for example `check.txtar` pipes input,
checks `stdout`, and asserts a missing file writes to `stderr`.

## The exit-code contract, tested on both sides

The `0/1/2/>2` contract (ADR-0002) is covered where each half is cheapest:

- **Exact codes** — the Go tests assert the precise integer: `0` and `2` through
  `cli.Run` (valid runs, usage errors), `1` and `3` through `runCmd` with a
  subcommand returning a typed `CodedError`.
- **End-to-end behaviour** — the `.txtar` scripts assert *success vs failure*
  (`exec` vs `! exec`) and that output lands on the right stream.

Together: the scripts prove the program behaves, the Go tests pin the numbers.

## Conventions

- **Pure stdlib.** Failures use `t.Errorf`/`t.Fatalf` with a `got/want` message.
  No third-party assertion libraries.
- **Inject, don't touch globals.** Config, logging, and progress take their
  streams and environment as parameters, so tests never mutate `os.Environ` or
  real files (use `t.TempDir`, `t.Setenv`).
- **Run with the race detector**: `go test -race ./...`.

The scaffold program itself is golden-tested the same way — see
[instantiating a new tool](instantiating-a-new-tool.md) and
`internal/scaffold/golden_test.go` in the repo.
