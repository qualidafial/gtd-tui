package projects

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

type keyMap struct {
	New        key.Binding
	Edit       key.Binding
	Enter      key.Binding
	Toggle     key.Binding
	Drop       key.Binding
	Park       key.Binding
	MoveUp     key.Binding
	MoveDown   key.Binding
	FocusQuery key.Binding

	nav     list.KeyMap
	editing help.KeyMap
}

func defaultKeyMap() keyMap {
	return keyMap{
		New:        key.NewBinding(key.WithKeys("+", "insert"), key.WithHelp("+/insert", "new project")),
		Edit:       key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view")),
		Toggle:     key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "complete")),
		Drop:       key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop")),
		Park:       key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "park")),
		MoveUp:     key.NewBinding(key.WithKeys("shift+up"), key.WithHelp("shift+↑", "move up")),
		MoveDown:   key.NewBinding(key.WithKeys("shift+down"), key.WithHelp("shift+↓", "move down")),
		FocusQuery: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	if k.editing != nil {
		return k.editing.ShortHelp()
	}
	return []key.Binding{
		k.nav.CursorUp,
		k.nav.CursorDown,
		k.FocusQuery,
		k.New,
		k.Edit,
		k.Enter,
		k.Toggle,
		k.Drop,
		k.Park,
		k.MoveUp,
		k.MoveDown,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	if k.editing != nil {
		return k.editing.FullHelp()
	}
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
			k.FocusQuery,
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