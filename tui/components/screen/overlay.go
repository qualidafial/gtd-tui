package screen

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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

func (o overlay) ShortHelp() []key.Binding {
	inner := o.inner.ShortHelp()
	if hasEsc(inner) {
		return inner
	}
	return slices.Concat(o.KeyMap.ShortHelp(), inner)
}

func (o overlay) FullHelp() [][]key.Binding {
	inner := o.inner.FullHelp()
	if slices.ContainsFunc(inner, hasEsc) {
		return inner
	}
	return slices.Concat(o.KeyMap.FullHelp(), inner)
}

func hasEsc(bindings []key.Binding) bool {
	return slices.ContainsFunc(bindings, func(b key.Binding) bool {
		return slices.Contains(b.Keys(), "esc")
	})
}

func (o overlay) CapturingInput() bool {
	return CapturingInput(o.inner)
}
