package tasklist

import (
	"context"
	"fmt"
	"slices"

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

// tasksReorderedMsg is delivered after a Shift+Up/Down move so the tab
// can refresh its items and keep the cursor on the moved task.
type tasksReorderedMsg struct {
	filter   gtd.TaskFilter
	tasks    []gtd.Task
	selectID int64
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
		tasks, err := m.svc.ListTasks(context.Background(), m.filter)
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
		// TasksLoadedMsg is broadcast to every tasklist tab; only apply it
		// when the loaded filter matches this tab's filter, otherwise an
		// Active-tab load lands in the Inbox tab and vice versa.
		if !filterMatches(msg.filter, m.filter) {
			return m, nil
		}
		items := make([]list.Item, len(msg.tasks))
		for i, t := range msg.tasks {
			items[i] = Item{t}
		}
		return m, m.list.SetItems(items)
	case tasksReorderedMsg:
		if !filterMatches(msg.filter, m.filter) {
			return m, nil
		}
		items := make([]list.Item, len(msg.tasks))
		idx := m.list.Index()
		for i, t := range msg.tasks {
			items[i] = Item{t}
			if t.ID == msg.selectID {
				idx = i
			}
		}
		cmd := m.list.SetItems(items)
		m.list.Select(idx)
		return m, cmd
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, KeyNew):
			t := gtd.Task{
				Status: gtd.TaskStatusPending,
				Kind:   gtd.TaskKindNextAction,
			}
			return m, screen.ShowOverlay(taskedit.New(t, m.svc))
		case key.Matches(msg, KeyEdit):
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.ShowOverlay(taskedit.New(ti.task, m.svc))
			}
		case key.Matches(msg, KeyDelete):
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.ShowOverlay(taskdelete.New(ti.task, m.svc))
			}
		case key.Matches(msg, KeyMoveUp):
			if cmd := m.moveCmd(-1); cmd != nil {
				return m, cmd
			}
		case key.Matches(msg, KeyMoveDown):
			if cmd := m.moveCmd(+1); cmd != nil {
				return m, cmd
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

func filterMatches(a, b gtd.TaskFilter) bool {
	statusMatch := (a.Status == nil && b.Status == nil) ||
		(a.Status != nil && b.Status != nil && *a.Status == *b.Status)
	return statusMatch && slices.Equal(a.TaskIDs, b.TaskIDs)
}

// moveCmd reorders the selected task by one slot in the given direction
// (-1 = up, +1 = down). Returns nil when the list is filtered, the
// status is closed, or no task is selected.
func (m Model) moveCmd(direction int) tea.Cmd {
	if m.list.FilterState() != list.Unfiltered {
		return nil
	}
	if isClosedFilter(m.filter) {
		return nil
	}
	cur, ok := m.list.SelectedItem().(Item)
	if !ok {
		return nil
	}

	id := cur.task.ID
	filter := m.filter
	svc := m.svc

	doMove := svc.MoveUp
	if direction > 0 {
		doMove = svc.MoveDown
	}

	return func() tea.Msg {
		ctx := context.Background()
		if err := doMove(ctx, id); err != nil {
			return fmt.Errorf("move task: %w", err)
		}
		tasks, err := svc.ListTasks(ctx, filter)
		if err != nil {
			return fmt.Errorf("reload tasks: %w", err)
		}
		return tasksReorderedMsg{
			filter:   filter,
			tasks:    tasks,
			selectID: id,
		}
	}
}

func isClosedFilter(f gtd.TaskFilter) bool {
	return f.Status != nil && (*f.Status == gtd.TaskStatusDone || *f.Status == gtd.TaskStatusDropped)
}
