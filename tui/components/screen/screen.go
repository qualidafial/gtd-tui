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

// ShowOverlayMsg signals a transition to a new screen.
type ShowOverlayMsg struct {
	Overlay Screen
}

func ShowOverlay(s Screen) tea.Cmd {
	return func() tea.Msg {
		return ShowOverlayMsg{Overlay: s}
	}
}

// HideOverlayMsg signals that the current overlay screen should be dismissed.
type HideOverlayMsg struct{}

func HideOverlay() tea.Cmd {
	return func() tea.Msg {
		return HideOverlayMsg{}
	}
}
