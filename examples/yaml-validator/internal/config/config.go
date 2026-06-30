// Package config loads CLI configuration with the precedence
//
//	flags > environment > config file > defaults
//
// This package resolves the lower three (env > file > defaults); the cli layer
// applies explicitly-set flags on top. No viper — a small stdlib + yaml.v3
// loader keeps the dependency surface lean (see the config ADR / ADR-0004).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the resolved, file/env-overridable settings. Fields are scalar so
// Config stays comparable. yaml tags drive file parsing; env suffixes are mapped
// explicitly in applyEnv.
type Config struct {
	Output    string `yaml:"output"`
	LogLevel  string `yaml:"log_level"`
	LogFormat string `yaml:"log_format"`
	NoColor   bool   `yaml:"no_color"`
}

// Defaults returns the built-in configuration, the lowest-precedence layer.
func Defaults() Config {
	return Config{
		Output:    "text",
		LogLevel:  "warn",
		LogFormat: "auto",
		NoColor:   false,
	}
}

// Source describes where to load config from. SearchDirs/ExplicitPath locate the
// file; Environ (os.Environ-style) supplies env overrides. All are injectable so
// Load is testable without touching the real filesystem or process environment.
type Source struct {
	AppName      string   // drives env prefix (APPNAME_) and file names
	ExplicitPath string   // from --config; when set it is the only file consulted
	SearchDirs   []string // searched for .<app>.yaml / config.yaml when no ExplicitPath
	Environ      []string // "KEY=VALUE" entries; nil means no env overrides
}

// Load resolves configuration applying defaults < file < env.
func Load(s Source) (Config, error) {
	cfg := Defaults()

	if path := findConfigFile(s); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return cfg, fmt.Errorf("reading config %s: %w", path, err)
		}
		// Unmarshal over the defaults so keys absent from the file keep their
		// default values.
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parsing config %s: %w", path, err)
		}
	}

	if err := applyEnv(&cfg, envPrefix(s.AppName), s.Environ); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// envPrefix derives the environment-variable prefix from the app name: uppercased
// with any non-alphanumeric run replaced by "_" (so "my-tool" -> "MY_TOOL_").
func envPrefix(appName string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(appName) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	b.WriteByte('_')
	return b.String()
}

// applyEnv overlays environment values (highest precedence below flags) onto cfg.
// Each field maps to <prefix><SUFFIX>; only present vars override.
func applyEnv(cfg *Config, prefix string, environ []string) error {
	env := map[string]string{}
	for _, kv := range environ {
		if k, v, ok := strings.Cut(kv, "="); ok {
			env[k] = v
		}
	}
	get := func(suffix string) (string, bool) {
		v, ok := env[prefix+suffix]
		return v, ok
	}

	if v, ok := get("OUTPUT"); ok {
		cfg.Output = v
	}
	if v, ok := get("LOG_LEVEL"); ok {
		cfg.LogLevel = v
	}
	if v, ok := get("LOG_FORMAT"); ok {
		cfg.LogFormat = v
	}
	if v, ok := get("NO_COLOR"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("invalid %sNO_COLOR=%q: %w", prefix, v, err)
		}
		cfg.NoColor = b
	}
	return nil
}

// findConfigFile returns the path to use, or "" if none. ExplicitPath (from
// --config) wins outright; otherwise each SearchDir is probed for ".<app>.yaml"
// then "config.yaml", first match wins.
func findConfigFile(s Source) string {
	if s.ExplicitPath != "" {
		return s.ExplicitPath
	}
	for _, dir := range s.SearchDirs {
		for _, name := range []string{"." + s.AppName + ".yaml", "config.yaml"} {
			p := filepath.Join(dir, name)
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return p
			}
		}
	}
	return ""
}
