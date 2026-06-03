package doitnow

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// KeyMap holds the do-it-now overlay's two bindings. Confirm marks the task
// done; Back leaves it open and dismisses.
type KeyMap struct {
	Confirm key.Binding
	Back    key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "done")),
		Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "leave open")),
	}
}

func (k KeyMap) Chords() []keymap.Group {
	return []keymap.Group{{
		{Binding: k.Confirm, Vis: keymap.Short},
		{Binding: k.Back, Vis: keymap.Short},
	}}
}
