package textfield

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	InsertNewline key.Binding
	MoveCursor    key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		InsertNewline: key.NewBinding(
			key.WithKeys("alt+enter", "ctrl+j"),
			key.WithHelp("alt+enter", "newline"),
		),
		MoveCursor: key.NewBinding(
			key.WithKeys("down", "ctrl+n", "up", "ctrl+p"),
			key.WithHelp("↑/↓", "move cursor"),
		),
	}
}

// Keys claims the textarea's newline and vertical cursor-movement keys
// so the form forwards them to the field (rather than inserting a newline
// becoming a no-op or up/down advancing fields). MoveCursor's hidden
// ctrl+n/ctrl+p aliases route but are not displayed.
func (m KeyMap) Keys() []keymap.Group {
	return []keymap.Group{{
		{Binding: m.InsertNewline, Show: []string{"alt+enter"}, Vis: keymap.Short},
		{Binding: m.MoveCursor, Show: []string{"up", "down"}, Vis: keymap.Short},
	}}
}
