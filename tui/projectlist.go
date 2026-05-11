package tui

import (
	tea "charm.land/bubbletea/v2"
)

type projectListScreen struct {
}

func newProjectListScreen() (projectListScreen, tea.Cmd) {
	return projectListScreen{}, nil
}

func (s projectListScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	return s, nil
}

func (s projectListScreen) View() string {
	return "Projects — coming soon."
}
