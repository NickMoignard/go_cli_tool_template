package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/NickMoignard/yamlvalidate/internal/logging"
)

func isJSON(s string) bool {
	var m map[string]any
	return json.Unmarshal([]byte(strings.TrimSpace(s)), &m) == nil
}

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
	}
	for name, want := range cases {
		if got := logging.ParseLevel(name); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestNew_JSONFormat_EmitsParseableJSON(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, logging.Options{Level: slog.LevelInfo, Format: logging.FormatJSON})

	log.Info("hello", "key", "val")

	if !isJSON(buf.String()) {
		t.Errorf("JSON format did not emit valid JSON:\n%s", buf.String())
	}
}

func TestNew_TextFormat_IsHumanConsoleNotJSON(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, logging.Options{Level: slog.LevelInfo, Format: logging.FormatText, Color: false})

	log.Info("plain message")

	out := buf.String()
	if !strings.Contains(out, "plain message") {
		t.Fatalf("missing message: %q", out)
	}
	if isJSON(out) {
		t.Errorf("text format emitted JSON, want human console output: %q", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("Color:false but output contains ANSI escapes: %q", out)
	}
}

func TestNew_TextFormat_Color_HasANSI(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, logging.Options{Level: slog.LevelInfo, Format: logging.FormatText, Color: true})

	log.Info("colored")

	if !strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("Color:true but output has no ANSI escapes: %q", buf.String())
	}
}

func TestNew_AutoFormat_ResolvesByTTY(t *testing.T) {
	var offTTY bytes.Buffer
	logging.New(&offTTY, logging.Options{Level: slog.LevelInfo, Format: logging.FormatAuto, IsTTY: false}).Info("x")
	if !isJSON(offTTY.String()) {
		t.Errorf("auto + non-TTY should be JSON, got: %q", offTTY.String())
	}

	var onTTY bytes.Buffer
	logging.New(&onTTY, logging.Options{Level: slog.LevelInfo, Format: logging.FormatAuto, IsTTY: true, Color: false}).Info("y")
	if isJSON(onTTY.String()) {
		t.Errorf("auto + TTY should be console text, got JSON: %q", onTTY.String())
	}
}

func TestNew_LevelFiltersBelowThreshold(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, logging.Options{Level: slog.LevelWarn, Format: logging.FormatJSON})

	log.Info("hidden info")
	log.Warn("shown warning")

	out := buf.String()
	if strings.Contains(out, "hidden info") {
		t.Errorf("output contains an info message below the warn threshold:\n%s", out)
	}
	if !strings.Contains(out, "shown warning") {
		t.Errorf("output missing the warning message:\n%s", out)
	}
}
