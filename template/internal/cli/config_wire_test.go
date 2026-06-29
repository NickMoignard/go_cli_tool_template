package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/OWNER/REPLACE_TOOL/internal/config"
)

func TestApplyConfig_FlagWinsConfigFillsRest(t *testing.T) {
	o := &globalOptions{Output: "json"} // pretend -o json was passed
	cfg := config.Config{Output: "text", LogLevel: "debug", LogFormat: "json", NoColor: true}
	// Only "output" was set on the command line.
	changed := func(name string) bool { return name == "output" }

	o.applyConfig(cfg, changed)

	if o.Output != "json" {
		t.Errorf("Output = %q, want json (flag wins over config)", o.Output)
	}
	if o.LogLevel != "debug" || o.LogFormat != "json" || !o.NoColor {
		t.Errorf("config should fill unset fields, got %+v", o)
	}
}

// End-to-end: --config file feeds the logger through the full chain.
func TestConfig_FileFlowsIntoLogger(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(cfgPath, []byte("log_level: debug\nlog_format: json\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use: "emit", SilenceErrors: true, SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loggerFrom(cmd.Context()).Info("info via config")
			return nil
		},
	})

	var stdout, stderr bytes.Buffer
	code := runCmd(context.Background(), root,
		[]string{"--config", cfgPath, "emit"},
		strings.NewReader(""), &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, ExitOK, stderr.String())
	}
	// log_level: debug means an Info log is shown (default warn would suppress it),
	// and log_format: json means it is JSON — proving config reached the logger.
	if !strings.Contains(stderr.String(), `"msg":"info via config"`) {
		t.Errorf("config (debug+json) did not flow into the logger; stderr=%q", stderr.String())
	}
}
