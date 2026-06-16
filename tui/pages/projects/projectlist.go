package projects

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
	"github.com/qualidafial/gtd-tui/internal/projectquery"
	"github.com/qualidafial/gtd-tui/tui/components/querybar"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectconvert"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectedit"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectstatus"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectview"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
)

const projectQueryDebounceDelay = 500 * time.Millisecond
const defaultProjectQuery = "status:open"

type Model struct {
	svc        gtd.ProjectService
	taskSvc    gtd.TaskService
	pickerFn   tasklist.PickerFactory
	taskViewFn tasklist.ViewFactory
	filter     gtd.ProjectFilter
	projects []gtd.Project
	query    querybar.Model
	list     list.Model
	KeyMap   KeyMap
	width    int
}

type projectsLoadedMsg struct {
	projects []gtd.Project
}

type taskCountsLoadedMsg struct {
	counts map[int64]gtd.ProjectTaskCounts
}

type projectsReorderedMsg struct {
	projects []gtd.Project
	selectID int64
}

func New(svc gtd.ProjectService, taskSvc gtd.TaskService, pickerFn tasklist.PickerFactory, taskViewFn tasklist.ViewFactory) Model {
	keys := DefaultKeyMap()
	l := list.New(nil, newDelegate(keys), 0, 0)
	l.SetStatusBarItemName("project", "projects")
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.Filter.SetEnabled(false)

	qb := querybar.New("/ ", "(all projects)", projectQueryDebounceDelay, func(s string) *querybar.ParseError {
		_, err := projectquery.Parse(s)
		if err == nil {
			return nil
		}
		if pe, ok := errors.AsType[*querybar.ParseError](err); ok {
			return pe
		}
		return &querybar.ParseError{Message: err.Error()}
	})
	qb.SetValue(defaultProjectQuery)

	filter, _ := projectquery.Parse(defaultProjectQuery)

	m := Model{
		svc:        svc,
		taskSvc:    taskSvc,
		pickerFn:   pickerFn,
		taskViewFn: taskViewFn,
		filter:     filter,
		query:      qb,
		list:       l,
		KeyMap:     keys,
	}
	m.updateKeybindings()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadCmd(), tea.RequestWindowSize)
}

func (m Model) loadCmd() tea.Cmd {
	svc := m.svc
	filter := m.filter
	return func() tea.Msg {
		projects, err := svc.ListProjects(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("load projects: %w", err)
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func (m Model) countCmd(projects []gtd.Project) tea.Cmd {
	if len(projects) == 0 {
		return nil
	}
	ids := make([]int64, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	svc := m.svc
	return func() tea.Msg {
		counts, err := svc.CountTasksByProjects(context.Background(), ids)
		if err != nil {
			return fmt.Errorf("count tasks: %w", err)
		}
		return taskCountsLoadedMsg{counts: counts}
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

	case projectsLoadedMsg:
		m.projects = msg.projects
		items := m.buildItems(nil)
		cmd := m.list.SetItems(items)
		m.updateKeybindings()
		return m, tea.Batch(cmd, m.countCmd(msg.projects))

	case taskCountsLoadedMsg:
		items := m.buildItems(msg.counts)
		cmd := m.list.SetItems(items)
		m.updateKeybindings()
		return m, cmd

	case projectsReorderedMsg:
		m.projects = msg.projects
		items := m.buildItems(m.currentCounts())
		idx := m.list.Index()
		for i, p := range msg.projects {
			if p.ID == msg.selectID {
				idx = i
				break
			}
		}
		cmd := m.list.SetItems(items)
		m.list.Select(idx)
		m.updateKeybindings()
		return m, cmd

	case projectconvert.ConfirmedMsg:
		return m, m.convertToTaskCmd(msg.ProjectID)

	case querybar.ApplyMsg:
		filter, err := projectquery.Parse(msg.Query)
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
			m.query.SetValue(defaultProjectQuery)
			filter, _ := projectquery.Parse(defaultProjectQuery)
			m.filter = filter
			return m, m.loadCmd()

		case key.Matches(msg, m.KeyMap.New):
			return m, screen.Push(projectedit.New(gtd.Project{}, m.svc, m.viewFactory()))

		case key.Matches(msg, m.KeyMap.Edit):
			if it, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(projectedit.New(it.project, m.svc, nil))
			}

		case key.Matches(msg, m.KeyMap.View):
			if it, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(projectview.New(it.project, m.taskSvc, m.svc, m.pickerFn, m.taskViewFn))
			}

		case key.Matches(msg, m.KeyMap.ConvertToTask):
			if it, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(projectconvert.New(it.project))
			}

		case key.Matches(msg, m.KeyMap.ToggleComplete):
			it, ok := m.list.SelectedItem().(Item)
			if !ok {
				break
			}
			switch it.project.Status {
			case gtd.ProjectStatusOpen:
				return m, screen.Push(projectstatus.New(it.project, it.counts.Total-it.counts.Complete, m.svc, projectstatus.Complete))
			default:
				return m, m.reopenCmd(it.project.ID)
			}

		case key.Matches(msg, m.KeyMap.Drop):
			if it, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(projectstatus.New(it.project, it.counts.Total-it.counts.Complete, m.svc, projectstatus.Drop))
			}

		case key.Matches(msg, m.KeyMap.Park):
			if it, ok := m.list.SelectedItem().(Item); ok {
				return m, m.parkCmd(it.project.ID)
			}

		case key.Matches(msg, m.KeyMap.MoveUp):
			if cmd := m.moveCmd(m.svc.MoveProjectUp); cmd != nil {
				return m, cmd
			}

		case key.Matches(msg, m.KeyMap.MoveDown):
			if cmd := m.moveCmd(m.svc.MoveProjectDown); cmd != nil {
				return m, cmd
			}

		case key.Matches(msg, m.KeyMap.MoveFirst):
			if cmd := m.moveCmd(m.svc.MoveProjectFirst); cmd != nil {
				return m, cmd
			}

		case key.Matches(msg, m.KeyMap.MoveLast):
			if cmd := m.moveCmd(m.svc.MoveProjectLast); cmd != nil {
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
	m.updateKeybindings()
	return m, cmd
}

func (m Model) reopenCmd(id int64) tea.Cmd {
	svc := m.svc
	filter := m.filter
	return func() tea.Msg {
		if _, err := svc.ReopenProject(context.Background(), id, time.Now()); err != nil {
			return fmt.Errorf("reopen project: %w", err)
		}
		projects, err := svc.ListProjects(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("reload projects: %w", err)
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func (m Model) convertToTaskCmd(id int64) tea.Cmd {
	svc := m.svc
	filter := m.filter
	return func() tea.Msg {
		if _, err := svc.ConvertProjectToTask(context.Background(), id); err != nil {
			return fmt.Errorf("convert project to task: %w", err)
		}
		projects, err := svc.ListProjects(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("reload projects: %w", err)
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func (m Model) parkCmd(id int64) tea.Cmd {
	svc := m.svc
	filter := m.filter
	return func() tea.Msg {
		if _, err := svc.ParkProject(context.Background(), id, time.Now()); err != nil {
			return fmt.Errorf("park project: %w", err)
		}
		projects, err := svc.ListProjects(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("reload projects: %w", err)
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func (m Model) moveCmd(doMove func(context.Context, int64, gtd.ProjectFilter) error) tea.Cmd {
	cur, ok := m.list.SelectedItem().(Item)
	if !ok {
		return nil
	}
	id := cur.project.ID
	svc := m.svc
	filter := m.filter
	return func() tea.Msg {
		ctx := context.Background()
		if err := doMove(ctx, id, filter); err != nil {
			return fmt.Errorf("move project: %w", err)
		}
		projects, err := svc.ListProjects(ctx, filter)
		if err != nil {
			return fmt.Errorf("reload projects: %w", err)
		}
		return projectsReorderedMsg{projects: projects, selectID: id}
	}
}

// buildItems constructs the list.Item slice from stored projects, merging in
// counts. A nil map leaves all counts at their zero value.
func (m Model) buildItems(counts map[int64]gtd.ProjectTaskCounts) []list.Item {
	items := make([]list.Item, len(m.projects))
	for i, p := range m.projects {
		items[i] = Item{project: p, counts: counts[p.ID]}
	}
	return items
}

// currentCounts extracts counts from the current list items so a reload can
// preserve them until fresh counts arrive.
func (m Model) currentCounts() map[int64]gtd.ProjectTaskCounts {
	items := m.list.Items()
	if len(items) == 0 {
		return nil
	}
	counts := make(map[int64]gtd.ProjectTaskCounts, len(items))
	for _, raw := range items {
		if it, ok := raw.(Item); ok {
			counts[it.project.ID] = it.counts
		}
	}
	return counts
}

func (m Model) viewFactory() projectedit.ViewFactory {
	return func(project gtd.Project) screen.Screen {
		return projectview.New(project, m.taskSvc, m.svc, m.pickerFn, m.taskViewFn)
	}
}

func (m *Model) updateKeybindings() {
	var status gtd.ProjectStatus
	var selectedItem Item
	selected := false
	if it, ok := m.list.SelectedItem().(Item); ok {
		status, selectedItem, selected = it.project.Status, it, true
	}

	open := selected && status == gtd.ProjectStatusOpen
	someday := selected && status == gtd.ProjectStatusSomeday
	orderable := open || someday

	// Convert-to-Task is offered only for an empty open project. counts.Total is
	// the non-dropped task count; a project with only dropped tasks is rare, and
	// the service guard rejects it with an error if it slips through.
	m.KeyMap.ConvertToTask.SetEnabled(selected && gtd.CanConvertProjectToTask(selectedItem.project, selectedItem.counts.Total))

	label := "toggle"
	switch {
	case !selected:
	case status == gtd.ProjectStatusOpen:
		label = "complete"
	default:
		label = "reopen"
	}
	m.KeyMap.ToggleComplete.SetHelp("space", label)
	m.KeyMap.Edit.SetEnabled(selected)
	m.KeyMap.View.SetEnabled(selected)
	m.KeyMap.ToggleComplete.SetEnabled(selected)
	m.KeyMap.Drop.SetEnabled(open || someday)
	m.KeyMap.Park.SetEnabled(open)

	// Revert-to-default is offered only when the live query differs from the
	// default; at the default it is a no-op and hidden.
	m.KeyMap.ResetQuery.SetEnabled(m.query.Value() != defaultProjectQuery)

	// Move-to-top tracks move-up; move-to-bottom tracks move-down. A reorder is
	// only valid within the selected project's status group, so the neighbor in
	// the move direction must itself be orderable.
	idx := m.list.Index()
	canMoveUp := orderable && idx > 0 && isOrderable(m.list, idx-1)
	canMoveDown := orderable && isOrderable(m.list, idx+1)
	m.KeyMap.MoveUp.SetEnabled(canMoveUp)
	m.KeyMap.MoveDown.SetEnabled(canMoveDown)
	m.KeyMap.MoveFirst.SetEnabled(canMoveUp)
	m.KeyMap.MoveLast.SetEnabled(canMoveDown)
}

// isOrderable reports whether the project at index i belongs to the reorderable
// set (open or someday).
func isOrderable(l list.Model, i int) bool {
	items := l.Items()
	if i < 0 || i >= len(items) {
		return false
	}
	it, ok := items[i].(Item)
	if !ok {
		return false
	}
	s := it.project.Status
	return s == gtd.ProjectStatusOpen || s == gtd.ProjectStatusSomeday
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

func (m Model) CapturingInput() bool { return m.query.CapturingInput() }

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.query.View())
	b.WriteByte('\n')
	b.WriteString(m.list.View())
	return b.String()
}
