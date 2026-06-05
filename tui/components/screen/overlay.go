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
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, o.KeyMap.Back) {
		if !CapturingInput(o.inner) {
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
