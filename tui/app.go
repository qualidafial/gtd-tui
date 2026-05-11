package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
)

var tabLabels = []string{"Inbox", "Tasks", "Projects", "Notes", "Timeline"}

// Model is the root bubbletea model. It owns the tab screens and an optional
// overlay screen (for edit views, etc.).
type Model struct {
	tabs      []Screen
	activeTab int
	overlay   Screen
	initCmd   tea.Cmd
	width     int
	height    int
}

func New(
	projectSvc gtd.ProjectService,
	taskSvc gtd.TaskService,
	projectTaskSvc gtd.ProjectTaskService,
) Model {
	inbox, inboxCmd := newInboxScreen(taskSvc)
	tasks, tasksCmd := newActiveTasksScreen(taskSvc)
	projects, projectsCmd := newProjectListScreen()
	notes, notesCmd := newNotesScreen()
	timeline, timelineCmd := newTimelineScreen()

	return Model{
		tabs:    []Screen{inbox, tasks, projects, notes, timeline},
		initCmd: tea.Batch(inboxCmd, tasksCmd, projectsCmd, notesCmd, timelineCmd),
	}
}

func (m Model) Init() tea.Cmd {
	return m.initCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case ChangeScreenMsg:
		m.overlay = msg.Screen
		return m, nil
	case CloseScreenMsg:
		m.overlay = nil
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.overlay == nil {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				return m, nil
			}
		case "shift+tab":
			if m.overlay == nil {
				m.activeTab = (m.activeTab + len(m.tabs) - 1) % len(m.tabs)
				return m, nil
			}
		}
	}

	if m.overlay != nil {
		var cmd tea.Cmd
		m.overlay, cmd = m.overlay.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.tabs[m.activeTab], cmd = m.tabs[m.activeTab].Update(msg)
	return m, cmd
}

var (
	logoStyle        = lipgloss.NewStyle().Bold(true).Padding(0, 1).Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	bulletStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	footerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m Model) View() tea.View {
	header := m.renderHeader()
	footer := m.renderFooter()

	var content string
	if m.overlay != nil {
		content = m.overlay.View()
	} else {
		content = m.tabs[m.activeTab].View()
	}

	v := tea.NewView(strings.Join([]string{header, content, footer}, "\n"))
	v.AltScreen = true
	return v
}

func (m Model) renderHeader() string {
	logo := logoStyle.Render("GTD")

	var tabs []string
	for i, label := range tabLabels {
		if i == m.activeTab {
			tabs = append(tabs, bulletStyle.Render("•")+" "+activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, "  "+inactiveTabStyle.Render(label))
		}
	}
	tabBar := strings.Join(tabs, "   ")

	return logo + "\n" + tabBar
}

func (m Model) renderFooter() string {
	keys := "tab/shift+tab: switch view  ctrl+c: quit"
	if m.overlay != nil {
		keys = "esc: back  ctrl+c: quit"
	}
	return footerStyle.Render(keys)
}
