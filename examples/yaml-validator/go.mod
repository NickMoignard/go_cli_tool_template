// Template module — the generic, compilable CLI skeleton the scaffold instantiates.
//
// The module path below is a SENTINEL token (ADR-0003): the scaffold replaces
// `github.com/NickMoignard/yamlvalidate` with the new project's real module path. Do not
// "fix" it to a real path — it is intentionally a placeholder.
module github.com/NickMoignard/yamlvalidate

go 1.25.0

require (
	github.com/lmittmann/tint v1.1.3
	github.com/rogpeppe/go-internal v1.15.0
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/schollz/progressbar/v3 v3.19.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/term v0.44.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
)
