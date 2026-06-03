package inbox

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// KeyMap holds the inbox screen's action bindings. New opens the capture
// overlay; Clarify (enter) opens the wizard on the selected item — items are
// write-once on capture, so no separate edit binding exists.
type KeyMap struct {
	New     key.Binding
	Clarify key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		New:     key.NewBinding(key.WithKeys("+", "insert"), key.WithHelp("+/insert", "new item")),
		Clarify: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "clarify")),
	}
}

func (k KeyMap) Chords() []keymap.Group {
	return []keymap.Group{{
		{Binding: k.New, Vis: keymap.Short},
		{Binding: k.Clarify, Vis: keymap.Short},
	}}
}
