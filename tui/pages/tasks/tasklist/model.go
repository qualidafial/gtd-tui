package tasklist

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/taskquery"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskedit"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

// queryAreaHeight is the fixed number of lines reserved above the list for the
// query bar, the offending-range underline, and the error message.
const queryAreaHeight = 3

// queryDebounceDelay is how long after the last keystroke a live validation
// parse fires.
const queryDebounceDelay = 2 * time.Second

var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

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
	filter        gtd.TaskFilter
	appliedQuery  string
	query         textinput.Model
	editing       bool
	parseErr      *taskquery.ParseError
	debounceSeq   int
	list          list.Model
	keys          keyMap
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

// queryDebounceMsg fires queryDebounceDelay after a keystroke; seq lets us
// ignore it when a newer keystroke has since arrived.
type queryDebounceMsg struct{ seq int }

func New(svc gtd.TaskService, query string, pickerFn PickerFactory, projectNameFn ProjectNameFunc) Model {
	keys := defaultKeyMap()

	l := list.New(nil, newDelegate(keys), 0, 0)
	l.SetStatusBarItemName("task", "tasks")
	l.SetShowTitle(false)
	l.SetShowHelp(false) // app renders help via mergedKeyMap

	l.DisableQuitKeybindings()
	// The query bar is the only filtering mechanism; disable the list's
	// built-in `/` filter so it doesn't compete.
	l.KeyMap.Filter.SetEnabled(false)

	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "(all tasks)"
	ti.SetValue(query)

	// Best-effort initial parse; an invalid seed query yields a zero filter.
	filter, _ := taskquery.Parse(query)

	m := Model{
		svc:           svc,
		pickerFn:      pickerFn,
		projectNameFn: projectNameFn,
		filter:        filter,
		appliedQuery:  query,
		query:         ti,
		list:          l,
		keys:          keys,
	}
	m.updateKeybindings()
	return m
}

func (m Model) resolveProjectName(task gtd.Task) string {
	if task.ProjectID == nil || m.projectNameFn == nil {
		return ""
	}
	return m.projectNameFn(*task.ProjectID)
}

func (m Model) Init() tea.Cmd {
	return m.loadCmd()
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
		listHeight := msg.Height - queryAreaHeight
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
	case queryDebounceMsg:
		if msg.seq == m.debounceSeq && m.editing {
			m.validate()
		}
		return m, nil
	case tea.KeyPressMsg:
		if m.editing {
			return m.updateEditing(msg)
		}
		switch {
		case key.Matches(msg, m.keys.FocusQuery):
			m.editing = true
			m.parseErr = nil
			return m, m.query.Focus()
		case key.Matches(msg, m.keys.New):
			t := gtd.Task{
				Status: gtd.TaskStatusOpen,
			}
			return m, screen.Push(taskedit.New(t, m.svc, ""))
		case key.Matches(msg, m.keys.Edit):
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(taskedit.New(ti.task, m.svc, m.resolveProjectName(ti.task)))
			}
		case key.Matches(msg, m.keys.Project):
			if ti, ok := m.list.SelectedItem().(Item); ok && m.pickerFn != nil {
				return m, screen.Push(m.pickerFn(ti.task))
			}
		case key.Matches(msg, m.keys.Toggle):
			if ti, ok := m.list.SelectedItem().(Item); ok {
				transition := taskstatus.Complete
				if ti.task.Status != gtd.TaskStatusOpen {
					transition = taskstatus.Reopen
				}
				return m, screen.Push(taskstatus.New(ti.task, m.svc, transition))
			}
		case key.Matches(msg, m.keys.Drop):
			// Drop is enabled only for pending tasks, so a match implies pending.
			if ti, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(taskstatus.New(ti.task, m.svc, taskstatus.Drop))
			}
		case key.Matches(msg, m.keys.MoveUp):
			if cmd := m.moveCmd(-1); cmd != nil {
				return m, cmd
			}
		case key.Matches(msg, m.keys.MoveDown):
			if cmd := m.moveCmd(+1); cmd != nil {
				return m, cmd
			}
		}
	}

	// While editing, non-key messages (cursor blink, etc.) go to the query bar.
	if m.editing {
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

// updateEditing handles key presses while the query bar is focused.
func (m Model) updateEditing(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Apply):
		filter, err := taskquery.Parse(m.query.Value())
		if err != nil {
			m.setParseErr(err)
			return m, nil // do not reload on parse failure
		}
		m.parseErr = nil
		m.editing = false
		m.query.Blur()
		m.appliedQuery = m.query.Value()
		m.filter = filter
		return m, m.loadCmd()
	case key.Matches(msg, m.keys.Cancel):
		m.editing = false
		m.query.Blur()
		m.query.SetValue(m.appliedQuery)
		m.parseErr = nil
		return m, nil
	}

	var cmd tea.Cmd
	m.query, cmd = m.query.Update(msg)
	m.debounceSeq++
	seq := m.debounceSeq
	tick := tea.Tick(queryDebounceDelay, func(time.Time) tea.Msg {
		return queryDebounceMsg{seq: seq}
	})
	return m, tea.Batch(cmd, tick)
}

// validate parses the current query for feedback only; it never reloads.
func (m *Model) validate() {
	if _, err := taskquery.Parse(m.query.Value()); err != nil {
		m.setParseErr(err)
	} else {
		m.parseErr = nil
	}
}

func (m *Model) setParseErr(err error) {
	if pe, ok := errors.AsType[*taskquery.ParseError](err); ok {
		m.parseErr = pe
	} else {
		m.parseErr = &taskquery.ParseError{Message: err.Error()}
	}
}

func (m Model) KeyMap() help.KeyMap {
	k := m.keys
	k.nav = m.list.KeyMap
	k.editing = m.editing
	return k
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
	m.keys.Toggle.SetHelp("space", label)
	m.keys.Project.SetEnabled(selected && m.pickerFn != nil)

	// Drop is valid only for pending tasks: the service rejects dropping a
	// done/dropped task (it must be reopened first).
	m.keys.Drop.SetEnabled(pending)

	// Reorder is limited to pending tasks, which sort above closed ones. Move
	// up is disabled on the first task; move down is disabled on the last
	// pending task (the next item is closed or there is none).
	idx := m.list.Index()
	m.keys.MoveUp.SetEnabled(pending && idx > 0)
	m.keys.MoveDown.SetEnabled(pending && statusAt(m.list, idx+1) == gtd.TaskStatusOpen)
}

// CapturingInput reports that the query bar is focused, so the app should not
// act on global keybindings (tab, help toggle) while the user is typing.
func (m Model) CapturingInput() bool {
	return m.editing
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.query.View())
	b.WriteByte('\n')

	if m.parseErr != nil {
		underline := strings.Repeat(" ", promptWidth(m.query)+m.parseErr.Start)
		n := m.parseErr.End - m.parseErr.Start
		if n < 1 {
			n = 1
		}
		underline += strings.Repeat("^", n)
		b.WriteString(errStyle.Render(underline))
		b.WriteByte('\n')
		b.WriteString(errStyle.Render(m.parseErr.Message))
		b.WriteByte('\n')
	} else {
		b.WriteString("\n\n")
	}

	b.WriteString(m.list.View())
	return b.String()
}

// promptWidth is the rune width of the query bar's prompt, used to align the
// offending-range underline under the input text.
func promptWidth(ti textinput.Model) int {
	return len([]rune(ti.Prompt))
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
			tasks:    tasks,
			selectID: id,
		}
	}
}
