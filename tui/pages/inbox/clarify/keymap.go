package clarify

import "charm.land/bubbles/v2/key"

// KeyMap holds the wizard's static keybindings. Per-step keys (y/n, m/s,
// t/s, enter) are matched directly in Update against these bindings so help
// can render them consistently.
type KeyMap struct {
	Yes     key.Binding
	No      key.Binding
	Me      key.Binding
	Someone key.Binding
	Trash   key.Binding
	Someday key.Binding
	Confirm key.Binding
	Back    key.Binding
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Yes:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
		No:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
		Me:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "me")),
		Someone: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "someone else")),
		Trash:   key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "trash")),
		Someday: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "someday")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Yes, k.No, k.Back}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Yes, k.No, k.Me, k.Someone, k.Trash, k.Someday, k.Confirm, k.Back}}
}
