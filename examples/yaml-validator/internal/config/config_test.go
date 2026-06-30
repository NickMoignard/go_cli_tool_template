package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NickMoignard/yamlvalidate/internal/config"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_NoFileNoEnv_ReturnsDefaults(t *testing.T) {
	got, err := config.Load(config.Source{AppName: "demo", SearchDirs: []string{t.TempDir()}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != config.Defaults() {
		t.Errorf("Load = %+v, want defaults %+v", got, config.Defaults())
	}
}

func TestLoad_ExplicitPathOverridesDiscovery(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".demo.yaml"), "output: json\n") // discoverable, should lose
	explicit := filepath.Join(dir, "custom.yaml")
	writeFile(t, explicit, "output: text\nlog_format: json\n")

	got, err := config.Load(config.Source{AppName: "demo", SearchDirs: []string{dir}, ExplicitPath: explicit})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Output != "text" || got.LogFormat != "json" {
		t.Errorf("Load = %+v, want explicit file values (output=text, log_format=json)", got)
	}
}

func TestLoad_ExplicitPathMissing_ReturnsError(t *testing.T) {
	// --config naming a missing file is an error (unlike optional discovery).
	_, err := config.Load(config.Source{AppName: "demo", ExplicitPath: filepath.Join(t.TempDir(), "nope.yaml")})
	if err == nil {
		t.Fatal("want an error when --config path does not exist")
	}
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".demo.yaml"), "output: [unclosed\n")

	if _, err := config.Load(config.Source{AppName: "demo", SearchDirs: []string{dir}}); err == nil {
		t.Fatal("want an error for malformed YAML")
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".demo.yaml"), "output: json\n")

	got, err := config.Load(config.Source{
		AppName:    "demo",
		SearchDirs: []string{dir},
		Environ:    []string{"DEMO_OUTPUT=text", "DEMO_NO_COLOR=true"},
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Output != "text" {
		t.Errorf("Output = %q, want text (env overrides file json)", got.Output)
	}
	if !got.NoColor {
		t.Errorf("NoColor = false, want true (DEMO_NO_COLOR=true)")
	}
}

func TestLoad_ReadsYAMLFileOverridingDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".demo.yaml"), "output: json\nlog_level: debug\n")

	got, err := config.Load(config.Source{AppName: "demo", SearchDirs: []string{dir}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Output != "json" {
		t.Errorf("Output = %q, want json (from file)", got.Output)
	}
	if got.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug (from file)", got.LogLevel)
	}
	if got.LogFormat != config.Defaults().LogFormat {
		t.Errorf("LogFormat = %q, want default %q (key absent from file)", got.LogFormat, config.Defaults().LogFormat)
	}
}
