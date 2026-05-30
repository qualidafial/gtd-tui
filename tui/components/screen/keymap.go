package screen

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Back key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Back}}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}
