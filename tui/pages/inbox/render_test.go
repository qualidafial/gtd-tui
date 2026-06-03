package inbox

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"

	"github.com/qualidafial/gtd-tui"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"single line", "hello world", "hello world"},
		{"newline", "line1\nline2", "line1 line2"},
		{"crlf", "line1\r\nline2", "line1 line2"},
		{"tabs", "a\tb", "a b"},
		{"collapses runs", "a  \n\n  b", "a b"},
		{"trims edges", "  a b  \n", "a b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flatten(tt.in); got != tt.want {
				t.Errorf("flatten(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestRenderSingleLine guards the help bar against multiline captures: a list
// row must occupy exactly one line regardless of newlines in the title or
// description, since the delegate reports Height() == 1.
func TestRenderSingleLine(t *testing.T) {
	items := []list.Item{
		Item{item: gtd.Item{
			Title:       "buy milk\nand eggs",
			Description: "from the\nstore\ndowntown",
		}},
	}
	d := newDelegate(DefaultKeyMap())
	m := list.New(items, d, 80, 10)

	var b strings.Builder
	d.Render(&b, m, 0, items[0])

	if got := strings.Count(b.String(), "\n"); got != 0 {
		t.Fatalf("rendered row has %d newlines, want 0: %q", got, b.String())
	}
}