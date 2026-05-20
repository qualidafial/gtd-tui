package tasklist

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskdelete"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskedit"
)

type Model struct {
	svc    gtd.TaskService
	filter gtd.TaskFilter
	list   list.Model
}

type TasksLoadedMsg struct {
	filter gtd.TaskFilter
	tasks  []gtd.Task
}

func NewInbox(svc gtd.TaskService) Model {
	return New(svc, gtd.TaskFilter{}.Status(gtd.TaskStatusInbox))
}

func NewActive(svc gtd.TaskService) Model {
	return New(svc, gtd.TaskFilter{}.Status(gtd.TaskStatusActive))
}

func New(svc gtd.TaskService, filter gtd.TaskFilter) Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.ShortHelpFunc = func() []key.Binding { return []key.Binding{KeyNew, KeyEdit} }
	delegate.FullHelpFunc = func() [][]key.Binding { return [][]key.Binding{{KeyNew, KeyEdit}} }

	l := list.New(nil, delegate, 0, 0)
	l.SetStatusBarItemName("task", "tasks")
	l.SetShowTitle(false)
	l.SetShowHelp(false) // app renders help via mergedKeyMap

	l.DisableQuitKeybindings()

	s := Model{
		svc:    svc,
		filter: filter,
		list:   l,
	}
	return s
}

func (m Model) Init() tea.Cmd {
	return m.loadCmd()
}

func (m Model) loadCmd() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.svc.Tasks(context.Background(), m.filter)
		if err != nil {
			return fmt.Errorf("load tasks: %w", err)
		}
		return TasksLoadedMsg{filter: m.filter, tasks: tasks}
	}
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tasks.TasksChangedMsg:
		return m, m.loadCmd()
	case TasksLoadedMsg:
		items := make([]list.Item, len(msg.tasks))
		for i, t := range msg.tasks {
			items[i] = Item{t}
		}
		return m, m.list.SetItems(items)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "n":
			var status gtd.TaskStatus = gtd.TaskStatusInbox
			if len(m.filter.Statuses) > 0 {
				status = m.filter.Statuses[0]
			}
			t := gtd.Task{
				Status: status,
			}
			return m, screen.ShowOverlay(taskedit.New(t, m.svc))
		case "enter":
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.ShowOverlay(taskedit.New(ti.task, m.svc))
			}
		case "delete":
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.ShowOverlay(taskdelete.New(ti.task, m.svc))
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) KeyMap() help.KeyMap {
	return KeyMap{km: m.list.KeyMap}
}

func (m Model) View() string {
	return m.list.View()
}
