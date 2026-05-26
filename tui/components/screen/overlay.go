package screen

import (
	"slices"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var keyEsc = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

type overlay struct {
	inner  Screen
	parent Screen
}

func Overlay(parent, child Screen) Screen {
	return overlay{inner: child, parent: parent}
}

func (o overlay) Pop() Screen {
	return o.parent
}

func (o overlay) Init() tea.Cmd {
	return o.inner.Init()
}

func (o overlay) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, keyEsc) {
		if !CapturingInput(o.inner) {
			return o, Dismiss()
		}
	}
	inner, cmd := o.inner.Update(msg)
	o.inner = inner
	return o, cmd
}

func (o overlay) View() string {
	return o.inner.View()
}

func (o overlay) KeyMap() help.KeyMap {
	return overlayKeyMap{inner: o.inner.KeyMap()}
}

func (o overlay) CapturingInput() bool {
	return CapturingInput(o.inner)
}

type overlayKeyMap struct {
	inner help.KeyMap
}

func (k overlayKeyMap) ShortHelp() []key.Binding {
	bindings := k.inner.ShortHelp()
	if hasEsc(bindings) {
		return bindings
	}
	return append(bindings, keyEsc)
}

func (k overlayKeyMap) FullHelp() [][]key.Binding {
	groups := k.inner.FullHelp()
	if slices.ContainsFunc(groups, hasEsc) {
		return groups
	}
	return append(groups, []key.Binding{keyEsc})
}

func hasEsc(bindings []key.Binding) bool {
	return slices.ContainsFunc(bindings, func(b key.Binding) bool {
		return slices.Contains(b.Keys(), "esc")
	})
}
