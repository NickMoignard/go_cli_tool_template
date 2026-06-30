# Guide: CI/CD for this repository

This guide covers the CI that runs on **this repository** — the one that builds
and tests the Template, the scaffold, and the worked example. It is distinct from
the CI a *scaffolded project* gets: that single-module workflow ships inside
`template/.github/workflows/ci.yml`, while everything here is the multi-module
workflow at the repo root.

## The core idea: every module, standalone

This repo is a three-module Go workspace (`go.work`), but the modules are built
and tested **independently with `GOWORK=off`** — exactly how they are consumed.
The scaffold copies `template/` *out* of the workspace, and a generated project
has no `go.work`, so CI verifying the standalone property is the whole point. The
workflow sets `GOWORK: "off"` globally and loops over `.`, `template`, and
`examples/yaml-validator`.

## Workflow: `.github/workflows/ci.yml`

Runs on pushes to `main`/`master` and on every pull request. The matrix is
deliberately **asymmetric** — full tests on Linux across two Go versions, a
cheaper build-smoke elsewhere:

| Job | Where | What |
|-----|-------|------|
| `test` | ubuntu × {stable, oldstable} Go | `gofmt` check, then `go test -race -covermode=atomic -coverprofile` per module, with a coverage total |
| `build` | macOS + Windows | `go build ./...` per module (compile-smoke on the other platforms) |
| `lint` | ubuntu (matrix per module) | `golangci-lint` (v2, pinned) |
| `govulncheck` | ubuntu (matrix per module) | `govulncheck ./...` on the latest stable toolchain |
| `example-drift` | ubuntu | re-scaffolds `template/` and diffs the non-domain files against `examples/yaml-validator/` |

A note on `govulncheck`: it only fails on a **called** vulnerability. Standard-
library advisories are resolved by CI's patched `stable` toolchain, so the job
stays green unless a real, reachable issue appears in your code or deps.

## Keeping the example honest: `example-drift`

`examples/yaml-validator/` is a hand-maintained instantiation of `template/` — the
scaffold output, then extended with the YAML-validation domain. As the Template
evolves, the example's **infrastructure** files (config, logging, progress,
version, the CLI plumbing, the shared txtar scripts) can silently fall behind what
a fresh scaffold would now emit. The `example-drift` job guards against that.

It runs [`scripts/check-example-drift.sh`](../../scripts/check-example-drift.sh),
which re-scaffolds `template/` into a temp dir with the example's exact identity
parameters (`-module github.com/NickMoignard/yamlvalidate -name yamlvalidate
-skip-git -skip-tidy`) and `diff -r`s the result against the committed example.
The diff **excludes the files the example legitimately owns as domain** — the
`validate` command and its tests/testdata, the rewritten `README.md`,
`go.mod`/`go.sum`, and the community/release/CI files the example deliberately
drops — so only genuine infrastructure divergence trips it. The exclude set lives
in the script, each entry commented with why it's domain.

Run it locally exactly as CI does:

```console
$ sh scripts/check-example-drift.sh
```

**Fixing a failure.** The output names every drifted file. For each one, decide:

- *The Template changed and the example should follow* (the usual case): port the
  change into `examples/yaml-validator/`, or regenerate the file from the current
  `template/`, then re-run the script until it's green.
- *The file became genuinely domain-specific*: add its basename to the `EXCLUDES`
  list in the script with a one-line comment justifying why the example owns it.

## Dependabot: `.github/dependabot.yml`

Weekly, grouped updates for Go modules in each of the three module directories
(`/`, `/template`, `/examples/yaml-validator`) plus the GitHub Actions used by the
workflows. Keeping `template/` updated means the scaffold emits fresh dependency
versions.

## Linting: `.golangci.yml`

The conservative default linter set (errcheck, govet, ineffassign, staticcheck,
unused) plus the `std-error-handling` exclusion preset, so the idiomatic unchecked
`fmt.Fprint*` calls to streams don't generate noise. It is discovered from each
module directory, so all three are linted with one ruleset. The Template ships its
own copy (`template/.golangci.yml`) so scaffolded projects lint clean too.

## Running it locally

There is no magic — reproduce CI with the same commands:

```console
$ gofmt -l .
$ for d in . template examples/yaml-validator; do
    ( cd "$d" && GOWORK=off go test -race ./... && GOWORK=off golangci-lint run ./... )
  done
```

## Releases

Tag-driven releases (GoReleaser) are documented separately — see
[Releasing & distribution](releasing.md). That pipeline ships inside `template/`
so scaffolded projects inherit it; this repo's CI is about keeping the Template
and its example green.
