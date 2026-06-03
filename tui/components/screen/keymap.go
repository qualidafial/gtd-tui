package screen

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	Back key.Binding
}

// Chords exposes the overlay's esc/back binding. When the inner subtree
// already claims esc, Resolve subtracts this chord so esc is not
// double-listed.
func (k KeyMap) Chords() []keymap.Group {
	return []keymap.Group{{{Binding: k.Back, Vis: keymap.Short}}}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}
