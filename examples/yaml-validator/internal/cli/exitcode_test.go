package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Exercises the full exit-code contract (ADR-0002) for the codes that need a
// command to produce them: a subcommand returning a typed CodedError must map to
// that code, with the message on stderr and stdout left clean. Codes 0 and 2 are
// covered end-to-end by the black-box cli.Run tests.
func TestRunCmd_MapsCodedErrorsToExitCodes(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"validation failure", validationErrorf("file %q does not conform", "x.yaml"), ExitInvalid},
		{"internal failure", internalErr(errors.New("disk on fire")), ExitInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCmd()
			root.AddCommand(&cobra.Command{
				Use:           "boom",
				SilenceErrors: true,
				SilenceUsage:  true,
				RunE:          func(*cobra.Command, []string) error { return tc.err },
			})

			var stdout, stderr bytes.Buffer
			code := runCmd(context.Background(), root, []string{"boom"}, strings.NewReader(""), &stdout, &stderr)

			if code != tc.want {
				t.Errorf("exit code = %d, want %d", code, tc.want)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if stderr.Len() == 0 {
				t.Error("stderr is empty, want the error message")
			}
		})
	}
}
