package doitnow

import "charm.land/bubbles/v2/key"

// KeyMap holds the do-it-now overlay's two bindings. Confirm marks the task
// done; Back leaves it open and dismisses.
type KeyMap struct {
	Confirm key.Binding
	Back    key.Binding
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "done")),
		Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "leave open")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Back}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Confirm, k.Back}}
}