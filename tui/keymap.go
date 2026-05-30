package tui

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Quit key.Binding
	Help key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

func (m KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		m.Quit,
	}
}

func (m KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.Help,
			m.Quit,
		},
	}
}
