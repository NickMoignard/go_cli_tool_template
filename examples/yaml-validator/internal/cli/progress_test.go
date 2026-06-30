package cli

import (
	"bytes"
	"testing"
)

func TestProgressEnabled(t *testing.T) {
	cases := []struct {
		name                  string
		isTTY, quiet, noInput bool
		want                  bool
	}{
		{"tty, interactive", true, false, false, true},
		{"non-tty", false, false, false, false},
		{"quiet suppresses", true, true, false, false},
		{"no-input suppresses", true, false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := progressEnabled(tc.isTTY, tc.quiet, tc.noInput); got != tc.want {
				t.Errorf("progressEnabled(%v,%v,%v) = %v, want %v", tc.isTTY, tc.quiet, tc.noInput, got, tc.want)
			}
		})
	}
}

func TestNewProgress_NonTTYWriter_IsSilent(t *testing.T) {
	o := &globalOptions{}
	var buf bytes.Buffer // not a terminal => progress must be suppressed

	tr := o.newProgress(&buf, 3, "working")
	tr.Add(1)
	tr.Finish()

	if buf.Len() != 0 {
		t.Errorf("non-TTY progress wrote %q, want silent", buf.String())
	}
}
