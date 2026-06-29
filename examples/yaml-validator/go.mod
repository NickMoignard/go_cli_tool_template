// Worked example: the Template instantiated and fleshed out into a real YAML
// (against JSON Schema) validator. It is its own module — as if it were a
// standalone scaffolded tool — so it proves the scaffold produces a releasable
// project and is built/tested by CI independently. See ADR-0001.
module github.com/NickMoignard/yamlvalidate

go 1.25
