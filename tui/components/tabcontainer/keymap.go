package tabcontainer

import "charm.land/bubbles/v2/key"

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

func (m KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{m.Next}
}

func (m KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{m.Next, m.Prev}}
}
