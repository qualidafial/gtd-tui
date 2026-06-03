package projectview

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	Edit key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
	}
}

func (k KeyMap) Chords() []keymap.Group {
	return []keymap.Group{{{Binding: k.Edit, Vis: keymap.Short}}}
}
