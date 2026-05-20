package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	gtd "github.com/qualidafial/gtd-tui"
)

type taskListScreen struct {
	svc    gtd.TaskService
	filter gtd.TaskFilter
	tasks  []gtd.Task
	cursor int
}

type tasksLoadedMsg struct {
	filter gtd.TaskFilter
	tasks  []gtd.Task
}

func newInboxScreen(svc gtd.TaskService) (taskListScreen, tea.Cmd) {
	return newTaskListScreen(svc, gtd.TaskFilter{}.Status(gtd.TaskStatusInbox))
}

func newActiveTasksScreen(svc gtd.TaskService) (taskListScreen, tea.Cmd) {
	return newTaskListScreen(svc, gtd.TaskFilter{}.Status(gtd.TaskStatusActive))
}

func newTaskListScreen(svc gtd.TaskService, filter gtd.TaskFilter) (taskListScreen, tea.Cmd) {
	s := taskListScreen{
		svc:    svc,
		filter: filter,
	}
	return s, s.loadCmd()
}

func (s taskListScreen) loadCmd() tea.Cmd {
	return func() tea.Msg {
		tasks, err := s.svc.Tasks(context.Background(), s.filter)
		if err != nil {
			return fmt.Errorf("load tasks: %w", err)
		}
		return tasksLoadedMsg{filter: s.filter, tasks: tasks}
	}
}

func (s taskListScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		// Only apply if the filter matches this screen instance.
		if filterKey(msg.filter) == filterKey(s.filter) {
			s.tasks = msg.tasks
			s.cursor = 0
		}
		return s, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if s.cursor < len(s.tasks)-1 {
				s.cursor++
			}
		case "k", "up":
			if s.cursor > 0 {
				s.cursor--
			}
		case "n":
			// TODO: open new task screen
		case "enter":
			// TODO: open task edit screen
		}
	}
	return s, nil
}

func (s taskListScreen) View() string {
	if len(s.tasks) == 0 {
		return "No tasks. Press 'n' to create one."
	}

	var out string
	for i, task := range s.tasks {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}
		out += fmt.Sprintf("%s%s\n", cursor, task.Title)
	}
	out += "\nj/k: navigate  enter: edit  n: new"
	return out
}

// filterKey produces a comparable key for a TaskFilter so loaded results can
// be routed to the correct screen instance.
func filterKey(f gtd.TaskFilter) string {
	parts := make([]string, len(f.Statuses))
	for i, s := range f.Statuses {
		parts[i] = string(s)
	}
	return strings.Join(parts, ",")
}
