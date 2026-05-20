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
