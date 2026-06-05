package tabcontainer

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	Next key.Binding
	Prev key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev view"),
		),
	}
}

// Keys exposes the view-switch bindings as one group. Next shows in both
// bars; Prev appears in full help only (matching the prior split).
func (m KeyMap) Keys() []keymap.Group {
	return []keymap.Group{{
		{Binding: m.Next, Vis: keymap.Short},
		{Binding: m.Prev, Vis: keymap.Full},
	}}
}
