package tasklist

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	"github.com/qualidafial/gtd-tui/tui/pages/tasks"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskdelete"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskedit"
)

// queryAreaHeight is the fixed number of lines reserved above the list for the
// query bar, the offending-range underline, and the error message.
const queryAreaHeight = 3

// queryDebounceDelay is how long after the last keystroke a live validation
// parse fires.
const queryDebounceDelay = 2 * time.Second

var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

type Model struct {
	svc          gtd.TaskService
	filter       gtd.TaskFilter
	appliedQuery string
	query        textinput.Model
	editing      bool
	parseErr     *taskquery.ParseError
	debounceSeq  int
	list         list.Model
	width        int
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

// queryDebounceMsg fires queryDebounceDelay after a keystroke; seq lets us
// ignore it when a newer keystroke has since arrived.
type queryDebounceMsg struct{ seq int }

func New(svc gtd.TaskService, query string) Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.ShortHelpFunc = func() []key.Binding { return []key.Binding{KeyNew, KeyEdit} }
	delegate.FullHelpFunc = func() [][]key.Binding { return [][]key.Binding{{KeyNew, KeyEdit}} }

	l := list.New(nil, delegate, 0, 0)
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

	return Model{
		svc:          svc,
		filter:       filter,
		appliedQuery: query,
		query:        ti,
		list:         l,
	}
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
		return TasksLoadedMsg{filter: filter, tasks: tasks}
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
		case key.Matches(msg, KeyFocusQuery):
			m.editing = true
			m.parseErr = nil
			return m, m.query.Focus()
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

	// While editing, non-key messages (cursor blink, etc.) go to the query bar.
	if m.editing {
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// updateEditing handles key presses while the query bar is focused.
func (m Model) updateEditing(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, KeyApply):
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
	case key.Matches(msg, KeyCancel):
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
	return KeyMap{km: m.list.KeyMap, editing: m.editing}
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

func filterMatches(a, b gtd.TaskFilter) bool {
	statusMatch := (a.Status == nil && b.Status == nil) ||
		(a.Status != nil && b.Status != nil && *a.Status == *b.Status)
	return statusMatch && slices.Equal(a.TaskIDs, b.TaskIDs)
}

// moveCmd reorders the selected task by one slot in the given direction
// (-1 = up, +1 = down). Returns nil when the status is closed or no task is
// selected.
func (m Model) moveCmd(direction int) tea.Cmd {
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
