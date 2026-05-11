package tui

import (
	tea "charm.land/bubbletea/v2"
)

type timelineScreen struct {
}

func newTimelineScreen() (timelineScreen, tea.Cmd) {
	return timelineScreen{}, nil
}

func (s timelineScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	return s, nil
}

func (s timelineScreen) View() string {
	return "Timeline — coming soon."
}
