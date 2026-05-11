package tui

import tea "charm.land/bubbletea/v2"

// Screen represents a full-screen view in the application.
type Screen interface {
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
}

// ChangeScreenMsg signals a transition to a new screen.
type ChangeScreenMsg struct {
	Screen Screen
}

func changeScreen(s Screen) tea.Cmd {
	return func() tea.Msg {
		return ChangeScreenMsg{Screen: s}
	}
}

// CloseScreenMsg signals that the current overlay screen should be dismissed.
type CloseScreenMsg struct{}

func closeScreen() tea.Cmd {
	return func() tea.Msg {
		return CloseScreenMsg{}
	}
}
