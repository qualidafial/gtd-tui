package projectview

import "charm.land/bubbles/v2/key"

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

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Edit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Edit}}
}
