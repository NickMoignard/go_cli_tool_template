package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Proves the whole chain: global flags → resolved options → logger construction
// → context → retrieval in a subcommand → output on stderr in the chosen format.
func TestLogger_WiredIntoContext_WritesJSONToStderr(t *testing.T) {
	root := NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use:           "emit",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loggerFrom(cmd.Context()).Warn("hello from subcommand")
			return nil
		},
	})

	var stdout, stderr bytes.Buffer
	code := runCmd(context.Background(), root,
		[]string{"--log-format", "json", "emit"},
		strings.NewReader(""), &stdout, &stderr)

	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d; stderr = %q", code, ExitOK, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("logs must go to stderr, not stdout; stdout = %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `"msg":"hello from subcommand"`) {
		t.Errorf("stderr missing the JSON log line; got %q", stderr.String())
	}
}
