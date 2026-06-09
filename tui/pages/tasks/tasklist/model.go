package tasklist

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/taskquery"
	"github.com/qualidafial/gtd-tui/tui/components/querybar"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskedit"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

// queryDebounceDelay is how long after the last keystroke a live validation
// parse fires.
const queryDebounceDelay = 500 * time.Millisecond

// PickerFactory creates a project picker overlay for the given task.
// When nil, the "p" keybinding is disabled.
type PickerFactory func(gtd.Task) screen.Screen

// ProjectNameFunc resolves a project ID to its display name.
// When nil, the task editor omits the project line.
type ProjectNameFunc func(id int64) string

type Model struct {
	svc           gtd.TaskService
	pickerFn      PickerFactory
	projectNameFn ProjectNameFunc
	defaultQuery  string
	filter        gtd.TaskFilter
	query         querybar.Model
	list          list.Model
	KeyMap        KeyMap
	width         int
}

type TasksLoadedMsg struct {
	tasks []gtd.Task
}

// tasksReorderedMsg is delivered after a Shift+Up/Down move so the tab
// can refresh its items and keep the cursor on the moved task.
type tasksReorderedMsg struct {
	tasks    []gtd.Task
	selectID int64
}

// New constructs a task list. showProjectChip controls whether each row
// renders a `+<project title>` chip for tasks that belong to a project; pass
// false for the in-project task list where every row would share the same
// project. The chip is independent of projectNameFn's role in the task editor.
func New(svc gtd.TaskService, query string, pickerFn PickerFactory, projectNameFn ProjectNameFunc, showProjectChip bool) Model {
	keys := DefaultKeyMap()

	l := list.New(nil, newDelegate(keys, projectChipResolver(showProjectChip, projectNameFn)), 0, 0)
	l.SetStatusBarItemName("task", "tasks")
	l.SetShowTitle(false)
	l.SetShowHelp(false) // app renders help via mergedKeyMap

	l.DisableQuitKeybindings()
	// The query bar is the only filtering mechanism; disable the list's
	// built-in `/` filter so it doesn't compete.
	l.KeyMap.Filter.SetEnabled(false)

	qb := querybar.New("/ ", "(all tasks)", queryDebounceDelay, func(s string) *querybar.ParseError {
		_, err := taskquery.Parse(s)
		if err == nil {
			return nil
		}
		if pe, ok := errors.AsType[*querybar.ParseError](err); ok {
			return pe
		}
		return &querybar.ParseError{Message: err.Error()}
	})
	qb.SetValue(query)

	// Best-effort initial parse; an invalid seed query yields a zero filter.
	filter, _ := taskquery.Parse(query)

	m := Model{
		svc:           svc,
		pickerFn:      pickerFn,
		projectNameFn: projectNameFn,
		defaultQuery:  query,
		filter:        filter,
		query:         qb,
		list:          l,
		KeyMap:        keys,
	}
	m.updateKeybindings()
	return m
}

// projectChipResolver returns a per-task project-name resolver suitable for
// the row renderer. It returns "" — suppressing the chip — when the list is
// configured without a project chip, when no resolver was supplied, or when
// the task is standalone.
func projectChipResolver(enabled bool, fn ProjectNameFunc) projectResolver {
	if !enabled || fn == nil {
		return nil
	}
	return func(t gtd.Task) string {
		if t.ProjectID == nil {
			return ""
		}
		return fn(*t.ProjectID)
	}
}

func (m Model) resolveProjectName(task gtd.Task) string {
	if task.ProjectID == nil || m.projectNameFn == nil {
		return ""
	}
	return m.projectNameFn(*task.ProjectID)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadCmd(), tea.RequestWindowSize)
}

func (m Model) loadCmd() tea.Cmd {
	filter := m.filter
	return func() tea.Msg {
		tasks, err := m.svc.ListTasks(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("load tasks: %w", err)
		}
		return TasksLoadedMsg{tasks: tasks}
	}
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.query.SetWidth(msg.Width)
		listHeight := msg.Height - 1
		if listHeight < 0 {
			listHeight = 0
		}
		m.list.SetSize(msg.Width, listHeight)
		return m, nil
	case TasksLoadedMsg:
		items := make([]list.Item, len(msg.tasks))
		for i, t := range msg.tasks {
			items[i] = Item{t}
		}
		cmd := m.list.SetItems(items)
		m.updateKeybindings()
		return m, cmd
	case tasksReorderedMsg:
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
		m.updateKeybindings()
		return m, cmd
	case querybar.ApplyMsg:
		filter, err := taskquery.Parse(msg.Query)
		if err != nil {
			return m, nil
		}
		m.filter = filter
		return m, m.loadCmd()
	case tea.KeyPressMsg:
		if m.query.CapturingInput() {
			var cmd tea.Cmd
			m.query, cmd = m.query.Update(msg)
			return m, cmd
		}
		switch {
		case key.Matches(msg, m.KeyMap.FocusQuery):
			var cmd tea.Cmd
			m.query, cmd = m.query.Focus()
			return m, cmd
		case key.Matches(msg, m.KeyMap.ResetQuery):
			m.query.SetValue(m.defaultQuery)
			filter, _ := taskquery.Parse(m.defaultQuery)
			m.filter = filter
			return m, m.loadCmd()
		case key.Matches(msg, m.KeyMap.New):
			t := gtd.Task{
				Status: gtd.TaskStatusOpen,
			}
			return m, screen.Push(taskedit.New(t, m.svc, ""))
		case key.Matches(msg, m.KeyMap.Edit):
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(taskedit.New(ti.task, m.svc, m.resolveProjectName(ti.task)))
			}
		case key.Matches(msg, m.KeyMap.Project):
			if ti, ok := m.list.SelectedItem().(Item); ok && m.pickerFn != nil {
				return m, screen.Push(m.pickerFn(ti.task))
			}
		case key.Matches(msg, m.KeyMap.ToggleComplete):
			if ti, ok := m.list.SelectedItem().(Item); ok {
				transition := taskstatus.Complete
				if ti.task.Status != gtd.TaskStatusOpen {
					transition = taskstatus.Reopen
				}
				return m, screen.Push(taskstatus.New(ti.task, m.svc, transition))
			}
		case key.Matches(msg, m.KeyMap.Drop):
			// Drop is enabled only for pending tasks, so a match implies pending.
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(taskstatus.New(ti.task, m.svc, taskstatus.Drop))
			}
		case key.Matches(msg, m.KeyMap.MoveUp):
			if cmd := m.moveCmd(-1); cmd != nil {
				return m, cmd
			}
		case key.Matches(msg, m.KeyMap.MoveDown):
			if cmd := m.moveCmd(+1); cmd != nil {
				return m, cmd
			}
		}
	}

	// While the query bar is capturing input, non-key messages go to it.
	if m.query.CapturingInput() {
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	// Navigation may have moved the cursor to a task with a different status;
	// refresh the per-selection binding state so help and key.Matches agree.
	m.updateKeybindings()
	return m, cmd
}

// Keys delegates to the query bar while it is capturing input;
// otherwise it contributes the action columns plus list-navigation and
// paging groups. Disabled action bindings are skipped by Resolve.
func (m Model) Keys() []keymap.Group {
	if m.query.CapturingInput() {
		return m.query.Keys()
	}
	return slices.Concat(
		m.KeyMap.Keys(),
		[]keymap.Group{
			{
				{Binding: m.list.KeyMap.CursorUp, Vis: keymap.Short},
				{Binding: m.list.KeyMap.CursorDown, Vis: keymap.Short},
				{Binding: m.list.KeyMap.GoToStart, Vis: keymap.Full},
				{Binding: m.list.KeyMap.GoToEnd, Vis: keymap.Full},
			},
			{
				{Binding: m.list.KeyMap.PrevPage, Vis: keymap.Full},
				{Binding: m.list.KeyMap.NextPage, Vis: keymap.Full},
			},
		},
	)
}

// updateKeybindings reconciles the per-selection action bindings with the
// currently-selected task. Call it after anything that changes the selection or
// the item set. Disabling a binding both hides it from the help bar and makes
// key.Matches reject it, so the bound action becomes inert.
func (m *Model) updateKeybindings() {
	status, selected := gtd.TaskStatus(""), false
	if ti, ok := m.list.SelectedItem().(Item); ok {
		status, selected = ti.task.Status, true
	}
	pending := selected && status == gtd.TaskStatusOpen

	label := "toggle"
	switch {
	case !selected:
	case status == gtd.TaskStatusOpen:
		label = "complete"
	default:
		label = "reopen"
	}
	m.KeyMap.ToggleComplete.SetHelp("space", label)
	m.KeyMap.Project.SetEnabled(selected && m.pickerFn != nil)

	// Revert-to-default is offered only when the live query differs from the
	// seed; at the default it is a no-op and hidden.
	m.KeyMap.ResetQuery.SetEnabled(m.query.Value() != m.defaultQuery)

	// Drop is valid only for pending tasks: the service rejects dropping a
	// done/dropped task (it must be reopened first).
	m.KeyMap.Drop.SetEnabled(pending)

	// Reorder is limited to pending tasks, which sort above closed ones. Move
	// up is disabled on the first task; move down is disabled on the last
	// pending task (the next item is closed or there is none).
	idx := m.list.Index()
	m.KeyMap.MoveUp.SetEnabled(pending && idx > 0)
	m.KeyMap.MoveDown.SetEnabled(pending && statusAt(m.list, idx+1) == gtd.TaskStatusOpen)
}

// CapturingInput reports that the query bar is focused, so the app should not
// act on global keybindings (tab, help toggle) while the user is typing.
func (m Model) CapturingInput() bool {
	return m.query.CapturingInput()
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.query.View())
	b.WriteByte('\n')
	b.WriteString(m.list.View())
	return b.String()
}

// statusAt returns the status of the task at index i, or the empty status if i
// is out of range or the item is not a task.
func statusAt(l list.Model, i int) gtd.TaskStatus {
	items := l.Items()
	if i < 0 || i >= len(items) {
		return ""
	}
	if it, ok := items[i].(Item); ok {
		return it.task.Status
	}
	return ""
}

// moveCmd reorders the selected task by one slot in the given direction
// (-1 = up, +1 = down). The move bindings are enabled only for a pending
// selection (see updateKeybindings), so reaching here implies a movable task;
// the guard below only covers an empty list.
func (m Model) moveCmd(direction int) tea.Cmd {
	cur, ok := m.list.SelectedItem().(Item)
	if !ok {
		return nil
	}

	id := cur.task.ID
	filter := m.filter
	svc := m.svc

	doMove := svc.MoveTaskUp
	if direction > 0 {
		doMove = svc.MoveTaskDown
	}

	return func() tea.Msg {
		ctx := context.Background()
		if err := doMove(ctx, id, filter); err != nil {
			return fmt.Errorf("move task: %w", err)
		}
		tasks, err := svc.ListTasks(ctx, filter)
		if err != nil {
			return fmt.Errorf("reload tasks: %w", err)
		}
		return tasksReorderedMsg{
			tasks:    tasks,
			selectID: id,
		}
	}
}
