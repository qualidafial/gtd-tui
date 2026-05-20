package taskdelete

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	keyBinds []key.Binding
}

func (m KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		m.keyBinds,
	}
}

func (m KeyMap) ShortHelp() []key.Binding {
	return m.keyBinds
}
