package check_test

import (
	"strings"
	"testing"

	"github.com/OWNER/REPLACE_TOOL/internal/check"
)

func TestCheck_EmptyInput_NotOK(t *testing.T) {
	got := check.Check("empty", strings.NewReader(""))

	if got.OK {
		t.Error("OK = true, want false for empty input")
	}
	if !strings.Contains(got.Problem, "empty") {
		t.Errorf("Problem = %q, want it to mention empty", got.Problem)
	}
}

func TestCheck_InvalidUTF8_NotOK(t *testing.T) {
	got := check.Check("binary", strings.NewReader("\xff\xfe\xfd"))

	if got.OK {
		t.Error("OK = true, want false for invalid UTF-8")
	}
	if !strings.Contains(got.Problem, "UTF-8") {
		t.Errorf("Problem = %q, want it to mention UTF-8", got.Problem)
	}
}

func TestCheck_ValidInput_OK(t *testing.T) {
	got := check.Check("greeting", strings.NewReader("hello, world\n"))

	if !got.OK {
		t.Errorf("OK = false (problem %q), want true for valid input", got.Problem)
	}
	if got.Name != "greeting" {
		t.Errorf("Name = %q, want %q", got.Name, "greeting")
	}
}
