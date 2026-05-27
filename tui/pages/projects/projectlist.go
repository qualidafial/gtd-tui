package projects

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectcreate"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectstatus"
)

var listErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

type Model struct {
	svc      gtd.ProjectService
	projects []gtd.Project
	list     list.Model
	keys     keyMap
	width    int
	err      error
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

func New(svc gtd.ProjectService) Model {
	keys := defaultKeyMap()
	l := list.New(nil, newDelegate(keys), 0, 0)
	l.SetStatusBarItemName("project", "projects")
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.Filter.SetEnabled(false)

	m := Model{
		svc:  svc,
		list: l,
		keys: keys,
	}
	m.updateKeybindings()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadCmd(), tea.RequestWindowSize)
}

func (m Model) loadCmd() tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		projects, err := svc.ListProjects(context.Background(), gtd.ProjectFilter{})
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
	case error:
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		// reserve one line at the bottom for the error footer
		m.list.SetSize(msg.Width, msg.Height-1)
		return m, nil

	case projectsLoadedMsg:
		m.err = nil
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
		m.err = nil
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

	case tea.KeyPressMsg:
		m.err = nil
		switch {
		case key.Matches(msg, m.keys.New):
			return m, screen.Push(projectcreate.New(m.svc))

		case key.Matches(msg, m.keys.Toggle):
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

		case key.Matches(msg, m.keys.Drop):
			if it, ok := m.list.SelectedItem().(Item); ok {
				return m, screen.Push(projectstatus.New(it.project, it.counts.Total-it.counts.Complete, m.svc, projectstatus.Drop))
			}

		case key.Matches(msg, m.keys.Park):
			if it, ok := m.list.SelectedItem().(Item); ok {
				return m, m.parkCmd(it.project.ID)
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

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	m.updateKeybindings()
	return m, cmd
}

func (m Model) reopenCmd(id int64) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		if _, err := svc.ReopenProject(context.Background(), id, time.Now()); err != nil {
			return fmt.Errorf("reopen project: %w", err)
		}
		projects, err := svc.ListProjects(context.Background(), gtd.ProjectFilter{})
		if err != nil {
			return fmt.Errorf("reload projects: %w", err)
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func (m Model) parkCmd(id int64) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		if _, err := svc.ParkProject(context.Background(), id, time.Now()); err != nil {
			return fmt.Errorf("park project: %w", err)
		}
		projects, err := svc.ListProjects(context.Background(), gtd.ProjectFilter{})
		if err != nil {
			return fmt.Errorf("reload projects: %w", err)
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func (m Model) moveCmd(direction int) tea.Cmd {
	cur, ok := m.list.SelectedItem().(Item)
	if !ok {
		return nil
	}
	id := cur.project.ID
	svc := m.svc
	doMove := svc.MoveProjectUp
	if direction > 0 {
		doMove = svc.MoveProjectDown
	}
	return func() tea.Msg {
		ctx := context.Background()
		if err := doMove(ctx, id); err != nil {
			return fmt.Errorf("move project: %w", err)
		}
		projects, err := svc.ListProjects(ctx, gtd.ProjectFilter{})
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

func (m *Model) updateKeybindings() {
	var status gtd.ProjectStatus
	selected := false
	if it, ok := m.list.SelectedItem().(Item); ok {
		status, selected = it.project.Status, true
	}

	open := selected && status == gtd.ProjectStatusOpen
	someday := selected && status == gtd.ProjectStatusSomeday
	orderable := open || someday

	label := "toggle"
	switch {
	case !selected:
	case status == gtd.ProjectStatusOpen:
		label = "complete"
	default:
		label = "reopen"
	}
	m.keys.Toggle.SetHelp("space", label)
	m.keys.Toggle.SetEnabled(selected)
	m.keys.Drop.SetEnabled(open || someday)
	m.keys.Park.SetEnabled(open)

	idx := m.list.Index()
	m.keys.MoveUp.SetEnabled(orderable && idx > 0 && isOrderable(m.list, idx-1))
	m.keys.MoveDown.SetEnabled(orderable && isOrderable(m.list, idx+1))
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

func (m Model) KeyMap() help.KeyMap {
	k := m.keys
	k.nav = m.list.KeyMap
	return k
}

func (m Model) CapturingInput() bool { return false }

func (m Model) View() string {
	footer := "\n" // blank reserved line
	if m.err != nil {
		footer = listErrStyle.Render(m.err.Error())
	}
	return m.list.View() + "\n" + footer
}
