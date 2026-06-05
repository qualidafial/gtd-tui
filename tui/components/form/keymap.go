package form

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// KeyMap is the navigation/submit keymap shared by every form. Cancel
// (Esc) is intentionally not owned by the form — overlays decide what
// cancellation means (dismiss, clear error, back out a wizard step) and
// surface their own Esc binding in the overlay's help.
type KeyMap struct {
	Next key.Binding
	Prev key.Binding
	Save key.Binding
}

// DefaultKeyMap returns the gtd-tui form keymap.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Next: key.NewBinding(key.WithKeys("tab", "down", "enter"), key.WithHelp("tab/↓", "next")),
		Prev: key.NewBinding(key.WithKeys("shift+tab", "up"), key.WithHelp("shift+tab/↑", "prev")),
		Save: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	}
}

// Keys exposes the form's navigation/submit bindings as a single group
// (one full-help column). Next routes tab/down/enter but shows only
// tab/↓ — enter is a hidden alias (advance/confirm on the last field) and
// down is routed here only when the focused field does not claim it.
func (m KeyMap) Keys() []keymap.Group {
	return []keymap.Group{{
		{Binding: m.Next, Show: []string{"tab", "down"}, Vis: keymap.Short},
		{Binding: m.Prev, Vis: keymap.Short},
		{Binding: m.Save, Vis: keymap.Short},
	}}
}
