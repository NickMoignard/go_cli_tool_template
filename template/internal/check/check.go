// Package check is the Template's placeholder domain logic. It is deliberately
// trivial — an input "passes" if it is non-empty and valid UTF-8 — but shaped
// exactly like real validation: read an input, return a verdict. To build a real
// tool, replace the body of Check (e.g. with JSON Schema validation, as the
// examples/yaml-validator does) while keeping this read→Result shape.
package check

import (
	"io"
	"unicode/utf8"
)

// Result is the outcome of checking one input.
type Result struct {
	Name    string `json:"name"`              // input name (a file path, or "<stdin>")
	OK      bool   `json:"ok"`                // whether the input passed
	Problem string `json:"problem,omitempty"` // reason when !OK; empty when OK
}

// Check reads all of r and applies the placeholder rule.
func Check(name string, r io.Reader) Result {
	data, err := io.ReadAll(r)
	if err != nil {
		return Result{Name: name, OK: false, Problem: "read error: " + err.Error()}
	}
	if len(data) == 0 {
		return Result{Name: name, OK: false, Problem: "input is empty"}
	}
	if !utf8.Valid(data) {
		return Result{Name: name, OK: false, Problem: "input is not valid UTF-8"}
	}
	return Result{Name: name, OK: true}
}
