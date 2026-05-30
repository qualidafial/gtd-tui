package querybar

import "charm.land/bubbles/v2/key"

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

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Apply, k.Cancel}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Apply, k.Cancel}}
}
