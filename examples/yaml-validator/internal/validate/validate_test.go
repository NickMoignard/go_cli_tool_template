package validate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NickMoignard/yamlvalidate/internal/validate"
)

const schemaJSON = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["name", "port"],
  "properties": {
    "name": {"type": "string"},
    "port": {"type": "integer", "minimum": 1, "maximum": 65535}
  },
  "additionalProperties": false
}`

func newValidator(t *testing.T, schema string) *validate.Validator {
	t.Helper()
	path := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(path, []byte(schema), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := validate.NewValidator(path)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	return v
}

func TestValidate_ConformingDoc_OK(t *testing.T) {
	v := newValidator(t, schemaJSON)
	res := v.Validate("good.yaml", strings.NewReader("name: web\nport: 8080\n"))
	if !res.OK {
		t.Errorf("OK = false, want true; violations: %+v", res.Violations)
	}
	if len(res.Violations) != 0 {
		t.Errorf("violations = %+v, want none", res.Violations)
	}
}

func TestValidate_MissingRequired_Violation(t *testing.T) {
	v := newValidator(t, schemaJSON)
	res := v.Validate("bad.yaml", strings.NewReader("name: web\n")) // no port
	if res.OK {
		t.Fatal("OK = true, want false for missing required field")
	}
	if len(res.Violations) == 0 {
		t.Fatal("want at least one violation")
	}
}

func TestValidate_WrongType_Violation(t *testing.T) {
	v := newValidator(t, schemaJSON)
	res := v.Validate("bad.yaml", strings.NewReader("name: web\nport: not-a-number\n"))
	if res.OK {
		t.Fatal("OK = true, want false for wrong type")
	}
	// The violation should point at the offending field.
	found := false
	for _, vio := range res.Violations {
		if strings.Contains(vio.Path, "port") {
			found = true
		}
		if vio.Message == "" {
			t.Errorf("violation has empty message: %+v", vio)
		}
	}
	if !found {
		t.Errorf("no violation referenced /port; got %+v", res.Violations)
	}
}

func TestValidate_OutOfRange_Violation(t *testing.T) {
	v := newValidator(t, schemaJSON)
	res := v.Validate("bad.yaml", strings.NewReader("name: web\nport: 99999\n"))
	if res.OK {
		t.Error("OK = true, want false for out-of-range integer")
	}
}

func TestValidate_AdditionalProperty_Violation(t *testing.T) {
	v := newValidator(t, schemaJSON)
	res := v.Validate("bad.yaml", strings.NewReader("name: web\nport: 80\nextra: nope\n"))
	if res.OK {
		t.Error("OK = true, want false for additional property")
	}
}

func TestValidate_MalformedYAML_NotOK(t *testing.T) {
	v := newValidator(t, schemaJSON)
	res := v.Validate("broken.yaml", strings.NewReader("name: : :\n  - bad indent\n: \n"))
	if res.OK {
		t.Error("OK = true, want false for malformed YAML")
	}
	if len(res.Violations) == 0 {
		t.Error("malformed YAML should yield a violation explaining the parse failure")
	}
}

func TestValidate_Name_IsCarried(t *testing.T) {
	v := newValidator(t, schemaJSON)
	res := v.Validate("svc.yaml", strings.NewReader("name: web\nport: 80\n"))
	if res.Name != "svc.yaml" {
		t.Errorf("Name = %q, want %q", res.Name, "svc.yaml")
	}
}

func TestNewValidator_MissingSchema_Error(t *testing.T) {
	if _, err := validate.NewValidator(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Error("NewValidator with missing schema: want error, got nil")
	}
}

func TestNewValidator_InvalidSchema_Error(t *testing.T) {
	if _, err := validate.NewValidator(writeTemp(t, "{ not valid json schema ")); err == nil {
		t.Error("NewValidator with invalid schema: want error, got nil")
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
