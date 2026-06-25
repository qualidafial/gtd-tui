package tasklist

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// KeyMap holds the task list's action bindings as a stable instance on the
// Model and also satisfies help.KeyMap. Bindings whose availability depends on
// the selected task (status, drop, move) are toggled in Model.updateKeybindings
// via SetEnabled. The status binding carries a fixed help label. Both help
// rendering and key.Matches honor the enabled flag, so a disabled binding is
// hidden from the help bar and inert when pressed — one source of truth for
// "is this action available now".
//
// The nav and editing fields are render-time context populated by Model.KeyMap
// just before the help component reads ShortHelp/FullHelp: nav mirrors the
// list's own (dynamically enabled) navigation bindings, and editing, when
// non-nil, delegates the advertised bindings to the query bar.
type KeyMap struct {
	New              key.Binding
	View             key.Binding
	Edit             key.Binding
	AssignToProject  key.Binding
	ConvertToProject key.Binding
	Status           key.Binding
	Drop             key.Binding
	MoveUp           key.Binding
	MoveDown         key.Binding
	MoveFirst        key.Binding
	MoveLast         key.Binding
	FocusQuery       key.Binding
	ResetQuery       key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		New:              key.NewBinding(key.WithKeys("c", "insert"), key.WithHelp("c", "new task")),
		View:             key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view")),
		Edit:             key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		AssignToProject:  key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "assign to project")),
		ConvertToProject: key.NewBinding(key.WithKeys("shift+c"), key.WithHelp("shift+c", "convert to project")),
		Status:           key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "status")),
		Drop:             key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop task")),
		MoveUp:           key.NewBinding(key.WithKeys("shift+up"), key.WithHelp("shift+↑", "move up")),
		MoveDown:         key.NewBinding(key.WithKeys("shift+down"), key.WithHelp("shift+↓", "move down")),
		MoveFirst:        key.NewBinding(key.WithKeys("shift+home"), key.WithHelp("shift+home", "move first")),
		MoveLast:         key.NewBinding(key.WithKeys("shift+end"), key.WithHelp("shift+end", "move last")),
		FocusQuery:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		ResetQuery:       key.NewBinding(key.WithKeys("\\"), key.WithHelp("\\", "reset filter")),
	}
}

// Keys returns every binding unconditionally as full-help columns; the
// help component and Resolve skip any that are disabled, so per-selection
// visibility is governed entirely by SetEnabled. Every binding shows in
// both bars (Vis Short).
func (k KeyMap) Keys() []keymap.Group {
	return []keymap.Group{
		{
			{Binding: k.FocusQuery, Vis: keymap.Short},
			{Binding: k.ResetQuery, Vis: keymap.Short},
		},
		{
			{Binding: k.New, Vis: keymap.Short},
			{Binding: k.View, Vis: keymap.Short},
			{Binding: k.Edit, Vis: keymap.Short},
			{Binding: k.AssignToProject, Vis: keymap.Short},
			{Binding: k.ConvertToProject, Vis: keymap.Short},
			{Binding: k.Status, Vis: keymap.Short},
			{Binding: k.Drop, Vis: keymap.Short},
		},
		{
			{Binding: k.MoveUp, Vis: keymap.Short},
			{Binding: k.MoveDown, Vis: keymap.Short},
			{Binding: k.MoveFirst, Vis: keymap.Short},
			{Binding: k.MoveLast, Vis: keymap.Short},
		},
	}
}
