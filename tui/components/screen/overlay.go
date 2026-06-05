package screen

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

type overlay struct {
	inner  Screen
	parent Screen
	KeyMap KeyMap
}

func Overlay(parent, child Screen) Screen {
	return overlay{
		inner:  child,
		parent: parent,
		KeyMap: DefaultKeyMap(),
	}
}

func (o overlay) Pop() Screen {
	return o.parent
}

func (o overlay) Init() tea.Cmd {
	return o.inner.Init()
}

func (o overlay) Update(msg tea.Msg) (Screen, tea.Cmd) {
	// esc is the overlay's fallback dismiss, but only when the inner subtree
	// does not itself claim it. Routing on keymap.Handles (the same binding
	// data Resolve dedups for help) lets an inner that owns esc — a modal
	// prompt recording its outcome, a focused query bar cancelling a filter —
	// handle the key, while inners that bind no esc still get the generic pop.
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, o.KeyMap.Back) {
		if !keymap.Handles(o.inner, msg) {
			return o, DismissCmd()
		}
	}
	inner, cmd := o.inner.Update(msg)
	o.inner = inner
	return o, cmd
}

func (o overlay) View() string {
	return o.inner.View()
}

// Keys aggregates the inner screen's full subtree (highest priority)
// ahead of the overlay's own esc binding. Resolve subtracts the overlay's
// esc when the inner subtree already claims it, so the previous bespoke
// hasEsc dedup is no longer needed.
func (o overlay) Keys() []keymap.Group {
	return slices.Concat(o.inner.Keys(), o.KeyMap.Keys())
}

func (o overlay) CapturingInput() bool {
	return CapturingInput(o.inner)
}
