---
name: scaffold-go-cli
description: Scaffold a new production-grade Go CLI project from the go_cli_tool_template.
disable-model-invocation: true
---

# Scaffold a Go CLI

Stand up a new Go CLI project from the
[`go_cli_tool_template`](https://github.com/NickMoignard/go_cli_tool_template) `template/`
module. Instantiation is **deterministic** and owned by `cmd/scaffold`: it copies the
template, substitutes the sentinel tokens, renames `cmd/<tool>`, and runs `go mod tidy`
+ `git init` (ADR-0003). Your job is thin — collect the inputs, bootstrap the scaffold,
invoke it, report. Never copy, rename, or substitute by hand: the script owns
correctness, and editing the output by hand reintroduces the drift it exists to prevent.

This skill is installed into arbitrary projects, so it does **not** assume it is running
inside the template repo. Step 2 fetches the scaffold for you.

## 1. Collect the inputs

Gather and confirm with the human:

- **module path** — the new project's Go module, e.g. `github.com/you/widget` (required)
- **tool name** — the binary and `cmd/` directory name, e.g. `widget` (required)
- **destination** — directory to create; must be empty or absent. **Use an absolute
  path** (step 2 runs from a cache dir, so a relative `-dest` would resolve there, not
  in the human's project). (required)
- **author** — copyright holder for LICENSE/README
- **description** — one line describing the tool
- **year** — defaults to the current year

Done when you hold the three required values; pass `author`/`description` empty only if
the human declines them.

## 2. Bootstrap the scaffold

The scaffold needs the template's `cmd/scaffold` + `template/` tree. Get it one of two
ways, in order of preference.

**A. A `scaffold` binary already on PATH.** If `command -v scaffold` succeeds, prefer it
— but only the self-contained build (template embedded) works standalone. Try it in
step 3; if it errors that it cannot find `template/`, fall back to B.

```console
$ command -v scaffold
```

**B. A cached clone of the template repo** (the default; needs only `git` + a Go
toolchain). Clone once into a cache dir and refresh on reuse:

```console
$ CACHE="${XDG_CACHE_HOME:-$HOME/.cache}/scaffold-go-cli"
$ REPO="$CACHE/go_cli_tool_template"
$ if [ -d "$REPO/.git" ]; then
    git -C "$REPO" pull --quiet --ff-only
  else
    mkdir -p "$CACHE"
    git clone --depth 1 https://github.com/NickMoignard/go_cli_tool_template.git "$REPO"
  fi
```

If neither a `scaffold` binary nor a Go toolchain (`command -v go`) is available, stop
and tell the human to install Go (or the `scaffold` binary) — the scaffold cannot run
without one.

## 3. Run the scaffold

Pass the inputs to the deterministic script — nothing more. With the cached clone (B),
run from the repo so `-source template` resolves, and give `-dest` as an **absolute**
path:

```console
$ ( cd "$REPO" && go run ./cmd/scaffold \
      -dest "<absolute destination>" \
      -module "<module path>" \
      -name "<tool name>" \
      -author "<author>" \
      -description "<description>" )
```

With a `scaffold` binary (A), the same flags apply, run from anywhere:

```console
$ scaffold -dest "<absolute destination>" -module "<module path>" -name "<tool name>" \
      -author "<author>" -description "<description>"
```

Then read the exit code:

- **0** — generated. Proceed.
- **2** — usage error (missing/invalid flag, or a non-empty destination). Fix the
  offending input and rerun.
- **1** — `go mod tidy` or `git init` failed (or, for path A, a non-embedded binary
  could not find `template/` — switch to B). Report the script's stderr verbatim; do
  not patch around it.

Done when the scaffold exits 0.

## 4. Report

Tell the human:

- where the project was created, and to `cd` into it;
- verify it with `go build ./cmd/<tool>` and `go test ./...` (both pass on fresh output);
- to build a real tool, replace the placeholder `internal/check` domain and its `check`
  command — the generated `README.md` and `docs/guides/instantiating-a-new-tool.md`
  walk through it.

Done when the human has the path, the verify commands, and the next edit.
