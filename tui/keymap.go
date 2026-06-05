package tui

import (
	"charm.land/bubbles/v2/key"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type KeyMap struct {
	Quit key.Binding
	Help key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// Keys exposes the global bindings as the app's lowest-priority group.
// Quit shows in both bars; Help appears in full help only (matching the
// prior ShortHelp/FullHelp split). Help is disabled while the active
// screen captures input, so it then claims and displays nothing.
func (m KeyMap) Keys() []keymap.Group {
	return []keymap.Group{{
		{Binding: m.Help, Vis: keymap.Full},
		{Binding: m.Quit, Vis: keymap.Short},
	}}
}
