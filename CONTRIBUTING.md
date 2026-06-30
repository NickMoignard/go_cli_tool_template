# Contributing to go_cli_tool_template

Thanks for your interest! This repo is a scaffold for Go CLI tools — it bundles a
compilable Template, the deterministic scaffold, and the guides. A few specifics
make contributing here different from a normal Go project.

## Layout & building

It is a three-module Go workspace (see the [README](README.md)). Each module is
built and tested **standalone**, the way it is actually consumed:

```console
$ go build ./...                                   # whole workspace
$ (cd template && GOWORK=off go test ./...)        # the Template, on its own
$ (cd examples/yaml-validator && GOWORK=off go test ./...)
```

The `template/` module uses sentinel placeholders (`REPLACE_TOOL`,
`github.com/OWNER/REPLACE_TOOL`, …) on purpose — do not "fix" them to real values.

## Before opening a pull request

Run what CI runs, per module:

```console
$ gofmt -l .
$ for d in . template examples/yaml-validator; do (cd "$d" && GOWORK=off go test -race ./... && GOWORK=off golangci-lint run ./...); done
```

If you change flags or commands, regenerate the help golden file:

```console
$ (cd template && GOWORK=off go test ./internal/cli/ -run TestRun_Help -update)
```

## Issue tracking

This project tracks work with **[beads](https://github.com/gastownhall/beads)**
(`bd`), not GitHub issues — see [CLAUDE.md](CLAUDE.md) and run `bd ready`. You are
still welcome to open a GitHub issue or pull request; a maintainer will reconcile
it with the beads graph.

## Architecture decisions

Significant decisions live in [`docs/adr/`](docs/adr/). If you propose a
load-bearing change, add or update an ADR alongside the code.

## Code of Conduct

By participating you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md).
