package querybar

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// KeyMap holds the bindings the query bar consumes when focused. It also
// satisfies help.KeyMap so parent screens can return it from their own KeyMap
// while editing, without knowing what bindings the query bar uses.
type KeyMap struct {
	Apply  key.Binding
	Cancel key.Binding
}

// DefaultKeyMap returns the standard query-bar bindings: enter applies, esc
// cancels.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Apply: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "apply"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

func (k KeyMap) Keys() []keymap.Group {
	return []keymap.Group{{
		{Binding: k.Apply, Vis: keymap.Short},
		{Binding: k.Cancel, Vis: keymap.Short},
	}}
}
