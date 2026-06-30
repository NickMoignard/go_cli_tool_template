package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NickMoignard/yamlvalidate/internal/cli"
)

const testSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["name", "port"],
  "properties": {
    "name": { "type": "string" },
    "port": { "type": "integer", "minimum": 1, "maximum": 65535 }
  },
  "additionalProperties": false
}`

// writeFile writes content to dir/name and returns the full path.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// schemaFile writes the shared test schema into a temp dir and returns its path.
func schemaFile(t *testing.T) string {
	t.Helper()
	return writeFile(t, t.TempDir(), "schema.json", testSchema)
}

// runIn is like run but feeds stdin, for exercising piped input.
func runIn(t *testing.T, stdin string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = cli.Run(context.Background(), args, strings.NewReader(stdin), &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func TestValidateCmd_ConformingFile_ExitOK(t *testing.T) {
	dir := t.TempDir()
	schema := writeFile(t, dir, "schema.json", testSchema)
	doc := writeFile(t, dir, "good.yaml", "name: web\nport: 8080\n")

	code, stdout, stderr := run(t, "validate", "--schema", schema, doc)

	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d (ExitOK); stderr = %q", code, cli.ExitOK, stderr)
	}
	if !strings.Contains(stdout, "ok") {
		t.Errorf("stdout = %q, want an ok report", stdout)
	}
}

func TestValidateCmd_NonConforming_ExitInvalid(t *testing.T) {
	dir := t.TempDir()
	schema := writeFile(t, dir, "schema.json", testSchema)
	doc := writeFile(t, dir, "bad.yaml", "name: web\nport: 99999\n")

	code, stdout, _ := run(t, "validate", "-s", schema, doc)

	if code != cli.ExitInvalid {
		t.Errorf("exit code = %d, want %d (ExitInvalid)", code, cli.ExitInvalid)
	}
	if !strings.Contains(stdout, "fail") {
		t.Errorf("stdout = %q, want a fail report", stdout)
	}
	if !strings.Contains(stdout, "port") {
		t.Errorf("stdout = %q, want the offending path reported", stdout)
	}
}

func TestValidateCmd_StdinPipe_NoArgs(t *testing.T) {
	schema := schemaFile(t)

	code, stdout, stderr := runIn(t, "name: web\nport: 80\n", "validate", "--schema", schema)

	if code != cli.ExitOK {
		t.Fatalf("exit code = %d, want %d; stderr = %q", code, cli.ExitOK, stderr)
	}
	if !strings.Contains(stdout, "<stdin>") {
		t.Errorf("stdout = %q, want it to name <stdin>", stdout)
	}
}

func TestValidateCmd_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	schema := writeFile(t, dir, "schema.json", testSchema)
	doc := writeFile(t, dir, "bad.yaml", "name: web\n") // missing port

	code, stdout, _ := run(t, "-o", "json", "validate", "-s", schema, doc)

	if code != cli.ExitInvalid {
		t.Fatalf("exit code = %d, want %d", code, cli.ExitInvalid)
	}
	var results []struct {
		Name       string `json:"name"`
		OK         bool   `json:"ok"`
		Violations []struct {
			Path    string `json:"path"`
			Message string `json:"message"`
		} `json:"violations"`
	}
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if len(results) != 1 || results[0].OK {
		t.Fatalf("results = %+v, want one failing result", results)
	}
	if len(results[0].Violations) == 0 {
		t.Errorf("want at least one violation in JSON output, got %+v", results[0])
	}
}

func TestValidateCmd_NoSchemaFlag_ExitUsage(t *testing.T) {
	doc := writeFile(t, t.TempDir(), "good.yaml", "name: web\nport: 80\n")

	code, stdout, stderr := run(t, "validate", doc)

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "schema") {
		t.Errorf("stderr = %q, want it to mention the missing schema", stderr)
	}
}

func TestValidateCmd_MissingSchemaFile_ExitUsage(t *testing.T) {
	doc := writeFile(t, t.TempDir(), "good.yaml", "name: web\nport: 80\n")
	missing := filepath.Join(t.TempDir(), "nope.json")

	code, _, stderr := run(t, "validate", "--schema", missing, doc)

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stderr == "" {
		t.Error("stderr = empty, want an explanation of the unusable schema")
	}
}

func TestValidateCmd_MissingInputFile_ExitUsage(t *testing.T) {
	schema := schemaFile(t)
	missing := filepath.Join(t.TempDir(), "nope.yaml")

	code, stdout, stderr := run(t, "validate", "--schema", schema, missing)

	if code != cli.ExitUsage {
		t.Errorf("exit code = %d, want %d (ExitUsage)", code, cli.ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "cannot open") {
		t.Errorf("stderr = %q, want it to report the open failure", stderr)
	}
}
