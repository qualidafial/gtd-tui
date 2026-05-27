package projectview

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(10)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type Model struct {
	project gtd.Project
	tasks   screen.Screen
	width   int
	height  int
}

func New(
	project gtd.Project,
	taskSvc gtd.TaskService,
	projectSvc gtd.ProjectService,
	pickerFn tasklist.PickerFactory,
) Model {
	wrapped := service.NewProjectTaskService(taskSvc, project.ID)
	tasks := tasklist.New(wrapped, "", pickerFn)

	return Model{
		project: project,
		tasks:   tasks,
	}
}

func (m Model) Init() tea.Cmd {
	return m.tasks.Init()
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
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
	}

	var cmd tea.Cmd
	m.tasks, cmd = m.tasks.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	header := m.renderHeader()
	return header + m.tasks.View()
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
	return strings.Join(lines, "\n") + "\n"
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

func (m Model) KeyMap() help.KeyMap {
	return m.tasks.KeyMap()
}
