package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/tabcontainer"
	"github.com/qualidafial/gtd-tui/tui/pages/projects"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectpicker"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
)

var (
	keyToggleHelp = key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help"))
	keyQuit       = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))
)

var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)
)

type Model struct {
	active screen.Screen
	help   help.Model
	width  int
	height int
}

func New(
	taskSvc gtd.TaskService,
	projectSvc gtd.ProjectService,
) Model {
	pickerFn := func(task gtd.Task) screen.Screen {
		return projectpicker.New(task, taskSvc, projectSvc)
	}
	pending := tasklist.New(taskSvc, "status:pending ready:now", pickerFn)
	projectList := projects.New(projectSvc, taskSvc, pickerFn)

	tabs := tabcontainer.New(
		tabcontainer.Tab{Label: "Tasks", Screen: pending},
		tabcontainer.Tab{Label: "Projects", Screen: projectList},
	)

	return Model{
		active: tabs,
		help:   help.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, tea.RequestWindowSize, m.active.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - appStyle.GetHorizontalMargins()
		m.height = msg.Height - appStyle.GetVerticalMargins()
		return m.resizeActive()
	case screen.PushMsg:
		m.active = screen.Overlay(m.active, msg.Screen)
		return m, tea.Batch(tea.RequestWindowSize, m.active.Init())
	case screen.DismissMsg:
		if popper, ok := m.active.(screen.Popper); ok {
			m.active = popper.Pop()
			return m, tea.Batch(tea.RequestWindowSize, m.active.Init())
		}
		return m, nil
	case screen.InitMsg:
		return m, tea.Batch(tea.RequestWindowSize, m.active.Init())
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keyQuit):
			return m, tea.Quit
		case key.Matches(msg, keyToggleHelp):
			if !screen.CapturingInput(m.active) {
				m.help.ShowAll = !m.help.ShowAll
				return m.resizeActive()
			}
		}
	}

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	return m, cmd
}

func (m Model) resizeActive() (tea.Model, tea.Cmd) {
	m.help.SetWidth(m.width)

	footer := m.renderFooter()
	innerHeight := m.height - lipgloss.Height(footer)

	msg := tea.WindowSizeMsg{
		Width:  m.width,
		Height: innerHeight,
	}

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	footer := m.renderFooter()

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.active.View(),
		footer,
	)

	content = appStyle.Render(content)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) renderFooter() string {
	km := appKeyMap{inner: m.active.KeyMap()}
	return "\n" + m.help.View(km)
}

type appKeyMap struct {
	inner help.KeyMap
}

func (k appKeyMap) ShortHelp() []key.Binding {
	return append(k.inner.ShortHelp(), keyQuit)
}

func (k appKeyMap) FullHelp() [][]key.Binding {
	groups := k.inner.FullHelp()
	return append(groups, []key.Binding{keyToggleHelp, keyQuit})
}
