package projectview

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	Edit          key.Binding
	LinkTask      key.Binding
	ConvertToTask key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		LinkTask: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "link task"),
		),
		ConvertToTask: key.NewBinding(
			key.WithKeys("shift+c"),
			key.WithHelp("shift+c", "convert to task"),
		),
	}
}

func (k KeyMap) Keys() []keymap.Group {
	return []keymap.Group{{
		{Binding: k.Edit, Vis: keymap.Short},
		{Binding: k.LinkTask, Vis: keymap.Short},
		{Binding: k.ConvertToTask, Vis: keymap.Short},
	}}
}
