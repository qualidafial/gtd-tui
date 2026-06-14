package projects

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	New            key.Binding
	Edit           key.Binding
	View           key.Binding
	ConvertToTask  key.Binding
	ToggleComplete key.Binding
	Drop           key.Binding
	Park           key.Binding
	MoveUp         key.Binding
	MoveDown       key.Binding
	FocusQuery     key.Binding
	ResetQuery     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		New:            key.NewBinding(key.WithKeys("+", "insert"), key.WithHelp("+/insert", "new project")),
		Edit:           key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		View:           key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view")),
		ConvertToTask:  key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "convert to task")),
		ToggleComplete: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "complete")),
		Drop:           key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop")),
		Park:           key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "park")),
		MoveUp:         key.NewBinding(key.WithKeys("shift+up"), key.WithHelp("shift+↑", "move up")),
		MoveDown:       key.NewBinding(key.WithKeys("shift+down"), key.WithHelp("shift+↓", "move down")),
		FocusQuery:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		ResetQuery:     key.NewBinding(key.WithKeys("\\"), key.WithHelp("\\", "reset filter")),
	}
}

// Keys exposes the project-list bindings as full-help columns; every
// binding shows in both bars (Vis Short). Per-selection availability is
// governed by SetEnabled, which Resolve and the help component honor.
func (k KeyMap) Keys() []keymap.Group {
	return []keymap.Group{
		{
			{Binding: k.FocusQuery, Vis: keymap.Short},
			{Binding: k.ResetQuery, Vis: keymap.Short},
		},
		{
			{Binding: k.New, Vis: keymap.Short},
			{Binding: k.Edit, Vis: keymap.Short},
			{Binding: k.View, Vis: keymap.Short},
			{Binding: k.ConvertToTask, Vis: keymap.Short},
		},
		{
			{Binding: k.ToggleComplete, Vis: keymap.Short},
			{Binding: k.Drop, Vis: keymap.Short},
			{Binding: k.Park, Vis: keymap.Short},
		},
		{
			{Binding: k.MoveUp, Vis: keymap.Short},
			{Binding: k.MoveDown, Vis: keymap.Short},
		},
	}
}
