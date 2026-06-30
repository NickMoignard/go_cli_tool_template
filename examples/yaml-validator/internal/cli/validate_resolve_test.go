package cli

import "testing"

func TestResolveInputNames(t *testing.T) {
	t.Run("no args on a TTY is a usage error", func(t *testing.T) {
		if _, err := resolveInputNames(nil, true); err == nil {
			t.Error("want a usage error when there is no input and stdin is a TTY")
		}
	})

	t.Run("no args with piped stdin reads stdin", func(t *testing.T) {
		names, err := resolveInputNames(nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 1 || names[0] != "-" {
			t.Errorf("names = %v, want [-] (stdin)", names)
		}
	})

	t.Run("args pass through", func(t *testing.T) {
		names, err := resolveInputNames([]string{"a.txt", "b.txt"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 2 {
			t.Errorf("names = %v, want the two args", names)
		}
	})
}
