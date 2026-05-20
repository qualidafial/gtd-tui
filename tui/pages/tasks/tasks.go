package tasks

import tea "charm.land/bubbletea/v2"

type TasksChangedMsg struct{}

func TasksChanged() tea.Cmd {
	return func() tea.Msg {
		return TasksChangedMsg{}
	}
}
