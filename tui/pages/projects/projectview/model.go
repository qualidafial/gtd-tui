package projectview

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectedit"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(10)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type Model struct {
	project    gtd.Project
	projectSvc gtd.ProjectService
	taskSvc    gtd.TaskService
	pickerFn   tasklist.PickerFactory
	tasks      screen.Screen
	KeyMap     KeyMap
	width      int
	height     int
}

func New(
	project gtd.Project,
	taskSvc gtd.TaskService,
	projectSvc gtd.ProjectService,
	pickerFn tasklist.PickerFactory,
) Model {
	wrapped := service.NewProjectTaskService(taskSvc, project.ID)
	projectNameFn := func(_ int64) string { return project.Title }
	tasks := tasklist.New(wrapped, "", pickerFn, projectNameFn, false)

	return Model{
		project:    project,
		projectSvc: projectSvc,
		taskSvc:    taskSvc,
		pickerFn:   pickerFn,
		tasks:      tasks,
		KeyMap:     DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.reloadCmd(), m.tasks.Init())
}

func (m Model) reloadCmd() tea.Cmd {
	if m.projectSvc == nil || m.project.ID == 0 {
		return nil
	}
	svc := m.projectSvc
	id := m.project.ID
	return func() tea.Msg {
		p, err := svc.GetProject(context.Background(), id)
		if err != nil {
			return fmt.Errorf("reload project: %w", err)
		}
		return projectReloadedMsg{project: p}
	}
}

type projectReloadedMsg struct {
	project gtd.Project
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case projectReloadedMsg:
		m.project = msg.project
		return m, tea.RequestWindowSize

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := lipgloss.Height(m.renderHeader())
		taskMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: msg.Height - headerHeight,
		}
		var cmd tea.Cmd
		m.tasks, cmd = m.tasks.Update(taskMsg)
		return m, cmd

	case tea.KeyPressMsg:
		if key.Matches(msg, m.KeyMap.Edit) && !screen.CapturingInput(m.tasks) {
			return m, screen.Push(projectedit.New(m.project, m.projectSvc, nil))
		}
	}

	var cmd tea.Cmd
	m.tasks, cmd = m.tasks.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.tasks.View(),
	)
}

func (m Model) renderHeader() string {
	var lines []string

	lines = append(lines, titleStyle.Render(m.project.Title))

	lines = append(lines, labelStyle.Render("Status:")+" "+statusStyle.Render(statusLabel(m.project.Status)))

	if m.project.Outcome != "" {
		lines = append(lines, labelStyle.Render("Outcome:")+" "+valueStyle.Render(m.project.Outcome))
	}

	if m.project.Due != nil {
		lines = append(lines, labelStyle.Render("Due:")+" "+valueStyle.Render(m.project.Due.Local().Format(time.DateOnly)))
	}

	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func statusLabel(s gtd.ProjectStatus) string {
	switch s {
	case gtd.ProjectStatusOpen:
		return "Open"
	case gtd.ProjectStatusSomeday:
		return "Someday"
	case gtd.ProjectStatusDone:
		return "Done"
	case gtd.ProjectStatusDropped:
		return "Dropped"
	default:
		return fmt.Sprintf("%s", s)
	}
}

func (m Model) CapturingInput() bool {
	return screen.CapturingInput(m.tasks)
}

func (m Model) Keys() []keymap.Group {
	return slices.Concat(m.KeyMap.Keys(), m.tasks.Keys())
}
