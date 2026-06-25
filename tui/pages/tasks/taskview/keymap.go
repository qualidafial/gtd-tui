package taskview

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	Edit             key.Binding
	Status           key.Binding
	Drop             key.Binding
	AssignToProject  key.Binding
	ConvertToProject key.Binding
	GoToProject      key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Edit:             key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Status:           key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "status")),
		Drop:             key.NewBinding(key.WithKeys("delete"), key.WithHelp("delete", "drop task")),
		AssignToProject:  key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "assign to project")),
		ConvertToProject: key.NewBinding(key.WithKeys("shift+c"), key.WithHelp("shift+c", "convert to project")),
		GoToProject:      key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "go to project")),
	}
}

// Keys returns every binding as a single help group; the help component and
// Resolve skip any binding disabled via SetEnabled, so per-task availability is
// governed entirely there.
func (k KeyMap) Keys() []keymap.Group {
	return []keymap.Group{{
		{Binding: k.Edit, Vis: keymap.Short},
		{Binding: k.Status, Vis: keymap.Short},
		{Binding: k.Drop, Vis: keymap.Short},
		{Binding: k.AssignToProject, Vis: keymap.Short},
		{Binding: k.ConvertToProject, Vis: keymap.Short},
		{Binding: k.GoToProject, Vis: keymap.Short},
	}}
}
