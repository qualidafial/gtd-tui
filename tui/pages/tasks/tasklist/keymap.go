package tasklist

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

var (
	KeyNew        = key.NewBinding(key.WithKeys("+", "insert"), key.WithHelp("+/insert", "new task"))
	KeyEdit       = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "edit"))
	KeyDelete     = key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop task"))
	KeyMoveUp     = key.NewBinding(key.WithKeys("shift+up"), key.WithHelp("shift+↑", "move up"))
	KeyMoveDown   = key.NewBinding(key.WithKeys("shift+down"), key.WithHelp("shift+↓", "move down"))
	KeyFocusQuery = key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter"))
	KeyApply      = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply"))
	KeyCancel     = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
)

type KeyMap struct {
	km      list.KeyMap
	editing bool
}

func (k KeyMap) ShortHelp() []key.Binding {
	if k.editing {
		return []key.Binding{KeyApply, KeyCancel}
	}
	return []key.Binding{
		k.km.CursorUp,
		k.km.CursorDown,
		KeyFocusQuery,
		KeyNew,
		KeyEdit,
		KeyDelete,
		KeyMoveUp,
		KeyMoveDown,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	if k.editing {
		return [][]key.Binding{{KeyApply, KeyCancel}}
	}
	return [][]key.Binding{
		{
			k.km.CursorUp,
			k.km.CursorDown,
			k.km.GoToStart,
			k.km.GoToEnd,
		},
		{
			k.km.PrevPage,
			k.km.NextPage,
		},
		{
			KeyFocusQuery,
		},
		{
			KeyNew,
			KeyEdit,
			KeyDelete,
		},
		{
			KeyMoveUp,
			KeyMoveDown,
		},
	}
}
