# yamlvalidate

A worked example for [`go_cli_tool_template`](../../): the template scaffolded into
a real tool that validates YAML documents against a [JSON Schema](https://json-schema.org/)
(Draft 2020-12, via [`santhosh-tekuri/jsonschema`](https://github.com/santhosh-tekuri/jsonschema)).

It exists to prove the scaffold produces a releasable, testable project. The
generic plumbing — global flags, layered config, structured logging, progress,
the exit-code contract — is exactly what the scaffold emits; the only domain
change is swapping the template's placeholder `check` command for the real
`validate` command in [`internal/validate`](internal/validate) and
[`internal/cli/validate.go`](internal/cli/validate.go).

## Usage

```console
$ yamlvalidate validate --schema schema.json config.yaml
```

Pass file paths, use `-` for stdin, or pipe data in. `--output json` emits a
machine-readable report; `--help` lists the global flags.

### Exit codes (ADR-0002)

| Code | Meaning                                              |
|------|------------------------------------------------------|
| 0    | every document conformed                             |
| 1    | a document was processed but failed validation       |
| 2    | usage error (missing/unreadable schema or input)     |
| >2   | unexpected internal error                            |

### Example

Given `schema.json`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["name", "port"],
  "properties": {
    "name": { "type": "string" },
    "port": { "type": "integer", "minimum": 1, "maximum": 65535 }
  },
  "additionalProperties": false
}
```

A conforming document passes:

```console
$ yamlvalidate validate --schema schema.json good.yaml
ok	good.yaml
```

A non-conforming one reports each violation by its JSON Pointer path and exits 1:

```console
$ yamlvalidate validate --schema schema.json bad.yaml
fail	bad.yaml
	/port: maximum: got 99,999, want 65,535
	/: additional properties 'extra' not allowed
$ echo $?
1
```

The same run with `--output json`:

```console
$ yamlvalidate -o json validate --schema schema.json bad.yaml
[
  {
    "name": "bad.yaml",
    "ok": false,
    "violations": [
      { "path": "/port", "message": "maximum: got 99,999, want 65,535" },
      { "path": "", "message": "additional properties 'extra' not allowed" }
    ]
  }
]
```

## Build & test

This module is built and tested on its own (outside the workspace), the same way
a freshly scaffolded project would be:

```console
$ GOWORK=off go test ./...
$ GOWORK=off go build ./cmd/yamlvalidate
```
