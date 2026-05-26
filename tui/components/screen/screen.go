package screen

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
)

// Screen represents a full-screen view in the application.
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
	KeyMap() help.KeyMap
}

// InputCapturer is an optional Screen capability. When a screen reports that it
// is capturing text input (e.g. a focused query bar), the app suppresses its
// global keybindings (tab, help toggle) so the keystrokes reach the screen.
type InputCapturer interface {
	CapturingInput() bool
}

// CapturingInput reports whether s is currently capturing text input.
func CapturingInput(s Screen) bool {
	c, ok := s.(InputCapturer)
	return ok && c.CapturingInput()
}

// Popper is satisfied by screens that can reveal a parent screen underneath.
type Popper interface {
	Pop() Screen
}

// PushMsg signals that a child screen should be pushed on top of the current view.
type PushMsg struct {
	Screen Screen
}

func Push(child Screen) tea.Cmd {
	return func() tea.Msg {
		return PushMsg{Screen: child}
	}
}

// DismissMsg signals that the current overlay screen should be dismissed.
type DismissMsg struct{}

func Dismiss() tea.Cmd {
	return func() tea.Msg {
		return DismissMsg{}
	}
}

// InitMsg signals that the active screen should re-initialize.
type InitMsg struct{}

func InitCmd() tea.Cmd {
	return func() tea.Msg {
		return InitMsg{}
	}
}