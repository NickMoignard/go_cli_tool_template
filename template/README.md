# REPLACE_TOOL

[![CI](https://github.com/OWNER/REPLACE_TOOL/actions/workflows/ci.yml/badge.svg)](https://github.com/OWNER/REPLACE_TOOL/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/OWNER/REPLACE_TOOL.svg)](https://pkg.go.dev/github.com/OWNER/REPLACE_TOOL)
[![Go Report Card](https://goreportcard.com/badge/github.com/OWNER/REPLACE_TOOL)](https://goreportcard.com/report/github.com/OWNER/REPLACE_TOOL)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

REPLACE_DESCRIPTION

## Install

```console
# Go toolchain
go install github.com/OWNER/REPLACE_TOOL/cmd/REPLACE_TOOL@latest

# Homebrew (replace OWNER with your GitHub username once the tap is published)
brew install OWNER/tap/REPLACE_TOOL
```

Or download a prebuilt archive for your platform from the
[latest release](https://github.com/OWNER/REPLACE_TOOL/releases/latest).

## Usage

```console
$ REPLACE_TOOL check input.txt
ok	input.txt
```

`REPLACE_TOOL` reads each input file (or stdin via `-` / a pipe), validates it,
and exits `0` if every input passes, `1` if any fails validation, or `2` for a
usage error. Use `-o json` for machine-readable output and `--help` for the full
flag set.

## Build from source

```console
$ git clone https://github.com/OWNER/REPLACE_TOOL
$ cd REPLACE_TOOL
$ go build ./cmd/REPLACE_TOOL
$ go test ./...
```

## Contributing

Issues and pull requests are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) and
the [Code of Conduct](CODE_OF_CONDUCT.md). To report a security issue, see
[SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © REPLACE_AUTHOR
