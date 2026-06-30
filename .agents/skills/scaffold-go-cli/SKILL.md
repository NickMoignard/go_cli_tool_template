---
name: scaffold-go-cli
description: Scaffold a new production-grade Go CLI project from this template.
disable-model-invocation: true
---

# Scaffold a Go CLI

Stand up a new Go CLI project from this repo's `template/` module. Instantiation is
**deterministic** and owned by `cmd/scaffold`: it copies the template, substitutes the
sentinel tokens, renames `cmd/<tool>`, and runs `go mod tidy` + `git init`. Your job is
thin — collect the inputs, invoke the script, report. Never copy, rename, or substitute
by hand: the script owns correctness (ADR-0003), and editing the output by hand
reintroduces the drift it exists to prevent.

## 1. Collect the inputs

Gather and confirm with the human:

- **module path** — the new project's Go module, e.g. `github.com/you/widget` (required)
- **tool name** — the binary and `cmd/` directory name, e.g. `widget` (required)
- **destination** — directory to create; must be empty or absent (required)
- **author** — copyright holder for LICENSE/README
- **description** — one line describing the tool
- **year** — defaults to the current year

Done when you hold the three required values; pass `author`/`description` empty only if the human declines them.

## 2. Run the scaffold

From the repository root, pass the inputs to the deterministic script — nothing more:

```console
$ go run ./cmd/scaffold \
    -dest "<destination>" \
    -module "<module path>" \
    -name "<tool name>" \
    -author "<author>" \
    -description "<description>"
```

Then read the exit code:

- **0** — generated. Proceed.
- **2** — usage error (missing/invalid flag, or a non-empty destination). Fix the offending input and rerun.
- **1** — `go mod tidy` or `git init` failed. Report the script's stderr verbatim; do not patch around it.

Done when the script exits 0.

## 3. Report

Tell the human:

- where the project was created, and to `cd` into it;
- verify it with `go build ./cmd/<tool>` and `go test ./...` (both pass on fresh output);
- to build a real tool, replace the placeholder `internal/check` domain and its `check`
  command — the generated `README.md` and `docs/guides/instantiating-a-new-tool.md`
  walk through it.

Done when the human has the path, the verify commands, and the next edit.
