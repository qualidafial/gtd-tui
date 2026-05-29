package tasklist

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

// keyMap holds the task list's action bindings as a stable instance on the
// Model and also satisfies help.KeyMap. Bindings whose availability depends on
// the selected task (drop, move) are toggled in Model.updateKeybindings via
// SetEnabled; the toggle binding's label is updated via SetHelp. Both help
// rendering and key.Matches honor the enabled flag, so a disabled binding is
// hidden from the help bar and inert when pressed — one source of truth for
// "is this action available now".
//
// The nav and editing fields are render-time context populated by Model.KeyMap
// just before the help component reads ShortHelp/FullHelp: nav mirrors the
// list's own (dynamically enabled) navigation bindings, and editing, when
// non-nil, delegates the advertised bindings to the query bar.
type keyMap struct {
	New        key.Binding
	Edit       key.Binding
	Project    key.Binding
	Toggle     key.Binding
	Drop       key.Binding
	MoveUp     key.Binding
	MoveDown   key.Binding
	FocusQuery key.Binding

	nav     list.KeyMap
	editing help.KeyMap
}

func defaultKeyMap() keyMap {
	return keyMap{
		New:        key.NewBinding(key.WithKeys("+", "insert"), key.WithHelp("+/insert", "new task")),
		Edit:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "edit")),
		Project:    key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "project")),
		Toggle:     key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "complete")),
		Drop:       key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop task")),
		MoveUp:     key.NewBinding(key.WithKeys("shift+up"), key.WithHelp("shift+↑", "move up")),
		MoveDown:   key.NewBinding(key.WithKeys("shift+down"), key.WithHelp("shift+↓", "move down")),
		FocusQuery: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	}
}

// ShortHelp and FullHelp return every binding unconditionally; the help
// component skips any that are disabled, so per-selection visibility is
// governed entirely by SetEnabled.
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
		k.Project,
		k.Toggle,
		k.Drop,
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
			k.Edit,
			k.Project,
			k.Toggle,
			k.Drop,
		},
		{
			k.MoveUp,
			k.MoveDown,
		},
	}
}
