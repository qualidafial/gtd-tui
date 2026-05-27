package projects

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

type keyMap struct {
	New      key.Binding
	Enter    key.Binding
	Toggle   key.Binding
	Drop     key.Binding
	Park     key.Binding
	MoveUp   key.Binding
	MoveDown key.Binding

	nav list.KeyMap
}

func defaultKeyMap() keyMap {
	return keyMap{
		New:      key.NewBinding(key.WithKeys("+", "insert"), key.WithHelp("+/insert", "new project")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view")),
		Toggle:   key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "complete")),
		Drop:     key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop")),
		Park:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "park")),
		MoveUp:   key.NewBinding(key.WithKeys("shift+up"), key.WithHelp("shift+↑", "move up")),
		MoveDown: key.NewBinding(key.WithKeys("shift+down"), key.WithHelp("shift+↓", "move down")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.nav.CursorUp,
		k.nav.CursorDown,
		k.New,
		k.Enter,
		k.Toggle,
		k.Drop,
		k.Park,
		k.MoveUp,
		k.MoveDown,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.nav.CursorUp,
			k.nav.CursorDown,
			k.nav.GoToStart,
			k.nav.GoToEnd,
		},
		{
			k.nav.PrevPage,
			k.nav.NextPage,
		},
		{
			k.New,
			k.Enter,
		},
		{
			k.Toggle,
			k.Drop,
			k.Park,
		},
		{
			k.MoveUp,
			k.MoveDown,
		},
	}
}