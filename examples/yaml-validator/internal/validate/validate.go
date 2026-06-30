// Package validate is the worked example's domain logic: it validates a YAML
// document against a JSON Schema (Draft 2020-12, via santhosh-tekuri/jsonschema).
//
// It mirrors the Template's check package shape — read an input, return a verdict
// — but with real validation and structured, path-addressed violations. That is
// the intended extension point: the scaffold produces the check placeholder; a
// real tool swaps in a package like this.
package validate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// Violation is one place where the document fails the schema.
type Violation struct {
	Path    string `json:"path"`    // JSON Pointer into the instance ("" = document root)
	Message string `json:"message"` // human-readable reason
}

// Result is the outcome of validating one document.
type Result struct {
	Name       string      `json:"name"`                 // input name (a file path, or "<stdin>")
	OK         bool        `json:"ok"`                   // whether the document conformed
	Violations []Violation `json:"violations,omitempty"` // why it failed; empty when OK
}

// Validator holds a compiled schema and validates documents against it.
type Validator struct {
	schema *jsonschema.Schema
}

// NewValidator compiles the JSON Schema at schemaPath. The schema may itself be
// written in JSON or YAML (YAML is a JSON superset). A read, parse, or schema
// compilation failure is returned as an error — the caller treats that as a usage
// error, since it means the supplied --schema is unusable.
func NewValidator(schemaPath string) (*Validator, error) {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	doc, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("parse schema %s: %w", schemaPath, err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaPath, doc); err != nil {
		return nil, fmt.Errorf("load schema %s: %w", schemaPath, err)
	}
	sch, err := c.Compile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("compile schema %s: %w", schemaPath, err)
	}
	return &Validator{schema: sch}, nil
}

// Validate reads a YAML document from r and validates it against the schema. It
// never returns an error: a malformed document or a schema failure is reported as
// a non-OK Result with violations, so the caller maps every outcome to the
// exit-code contract uniformly.
func (v *Validator) Validate(name string, r io.Reader) Result {
	data, err := io.ReadAll(r)
	if err != nil {
		return failed(name, "", "read error: "+err.Error())
	}
	inst, err := decode(data)
	if err != nil {
		return failed(name, "", "invalid YAML: "+err.Error())
	}
	if err := v.schema.Validate(inst); err != nil {
		var verr *jsonschema.ValidationError
		if errors.As(err, &verr) {
			return Result{Name: name, OK: false, Violations: violations(verr)}
		}
		return failed(name, "", err.Error())
	}
	return Result{Name: name, OK: true}
}

// violations flattens the schema validator's basic output into a list of leaf
// failures, each addressed by its instance location (a JSON Pointer).
func violations(verr *jsonschema.ValidationError) []Violation {
	var out []Violation
	var walk func(u *jsonschema.OutputUnit)
	walk = func(u *jsonschema.OutputUnit) {
		if u.Error != nil {
			out = append(out, Violation{Path: u.InstanceLocation, Message: u.Error.String()})
		}
		for i := range u.Errors {
			walk(&u.Errors[i])
		}
	}
	walk(verr.BasicOutput())
	if len(out) == 0 {
		// Defensive: always surface at least the top-level message.
		out = append(out, Violation{Path: "", Message: verr.Error()})
	}
	return out
}

// decode parses YAML/JSON bytes into the canonical value form the schema
// validator expects (numbers as json.Number, maps keyed by string). YAML decodes
// to JSON-compatible Go values, which we round-trip through JSON to normalize.
func decode(data []byte) (any, error) {
	var y any
	if err := yaml.Unmarshal(data, &y); err != nil {
		return nil, err
	}
	jb, err := json.Marshal(y)
	if err != nil {
		return nil, err
	}
	return jsonschema.UnmarshalJSON(bytes.NewReader(jb))
}

func failed(name, path, msg string) Result {
	return Result{Name: name, OK: false, Violations: []Violation{{Path: path, Message: msg}}}
}
