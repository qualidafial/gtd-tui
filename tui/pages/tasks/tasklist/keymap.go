package tasklist

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

var (
	KeyNew    = key.NewBinding(key.WithKeys("+", "insert"), key.WithHelp("+/insert", "new task"))
	KeyEdit   = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "edit"))
	KeyDelete = key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop task"))
)

type KeyMap struct {
	km list.KeyMap
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.km.CursorUp,
		k.km.CursorDown,
		KeyNew,
		KeyEdit,
		KeyDelete,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
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
			k.km.Filter,
			k.km.ClearFilter,
		},
		{
			KeyNew,
			KeyEdit,
			KeyDelete,
		},
	}
}
