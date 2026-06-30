# Contributing to REPLACE_TOOL

Thanks for your interest in improving REPLACE_TOOL! This project welcomes issues
and pull requests.

## Getting started

```console
$ git clone https://github.com/OWNER/REPLACE_TOOL
$ cd REPLACE_TOOL
$ go test ./...
```

## Before you open a pull request

Run the same checks CI does:

```console
$ gofmt -l .            # should print nothing
$ go vet ./...
$ go test -race ./...
$ golangci-lint run ./...
```

Please:

- Keep changes focused; one logical change per pull request.
- Add or update tests for any behaviour change.
- Write a clear commit message describing the *why*, not just the *what*.
- Follow the existing code style and conventions.

## Reporting bugs and requesting features

Open an issue using the appropriate template. For bugs, include the version
(`REPLACE_TOOL --version`), your OS/arch, and the smallest steps that reproduce
the problem.

## Code of Conduct

By participating you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md).
