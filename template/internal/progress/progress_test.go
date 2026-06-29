package progress_test

import (
	"bytes"
	"testing"

	"github.com/OWNER/REPLACE_TOOL/internal/progress"
)

func TestNew_Enabled_WritesProgressToWriter(t *testing.T) {
	var buf bytes.Buffer
	tr := progress.New(&buf, progress.Options{Enabled: true, Total: 3, Description: "working"})

	tr.Add(1)
	tr.Finish()

	if buf.Len() == 0 {
		t.Error("enabled tracker wrote nothing, want rendered progress output")
	}
}

func TestNew_Disabled_IsSilent(t *testing.T) {
	var buf bytes.Buffer
	tr := progress.New(&buf, progress.Options{Enabled: false, Total: 3})

	tr.Add(1)
	tr.Add(2)
	tr.Finish()

	if buf.Len() != 0 {
		t.Errorf("disabled tracker wrote %q, want nothing", buf.String())
	}
}
