package tui

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
)

// app-level key bindings
var (
	keyTab        = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next view"))
	keyShiftTab   = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev view"))
	keyToggleHelp = key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help"))
	keyQuit       = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))
)

var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)
)

var tabLabels = []string{
	"Tasks",
	// "Inbox",
	// "Projects",
	// "Notes",
	// "Timeline",
}

// Model is the root bubbletea model. It owns the tab screens and an optional
// overlay screen (for edit views, etc.).
type Model struct {
	tabs      []screen.Screen
	activeTab int
	overlay   screen.Screen
	initCmd   tea.Cmd
	help      help.Model
	width     int
	height    int
}

func New(
	// projectSvc gtd.ProjectService,
	taskSvc gtd.TaskService,
	// projectTaskSvc gtd.ProjectTaskService,
) Model {
	pending := tasklist.New(taskSvc, gtd.TaskFilter{}.WithStatus(gtd.TaskStatusPending))
	// inbox := tasklist.New(inboxSvc, ...)
	// projects, projectsCmd := newProjectListScreen()
	// notes, notesCmd := newNotesScreen()
	// timeline, timelineCmd := newTimelineScreen()

	screens := []screen.Screen{
		pending,
		// inbox,
		// projects,
		// notes,
		// timeline,
	}

	return Model{
		tabs: screens,
		help: help.New(),
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, screen := range m.tabs {
		cmds = append(cmds, screen.Init())
	}
	if m.overlay != nil {
		cmds = append(cmds, m.overlay.Init())
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - appStyle.GetHorizontalMargins()
		m.height = msg.Height - appStyle.GetVerticalMargins()

		return m.resizeScreens()
	case screen.ShowOverlayMsg:
		m.overlay = msg.Overlay
		return m, m.overlay.Init()
	case screen.HideOverlayMsg:
		m.overlay = nil
		return m, nil
	case tasks.TasksChangedMsg, tasklist.TasksLoadedMsg:
		// Broadcast cross-tab task events so every tasklist tab gets a
		// chance to react. Each tab filters TasksLoadedMsg by its own
		// filter, so messages addressed to other tabs are ignored.
		var cmds []tea.Cmd
		var cmd tea.Cmd
		for i := range m.tabs {
			m.tabs[i], cmd = m.tabs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keyQuit):
			return m, tea.Quit
		case key.Matches(msg, keyTab):
			if m.overlay == nil {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				return m, m.tabs[m.activeTab].Init()
			}
		case key.Matches(msg, keyShiftTab):
			if m.overlay == nil {
				m.activeTab = (m.activeTab + len(m.tabs) - 1) % len(m.tabs)
				return m, nil
			}
		case key.Matches(msg, keyToggleHelp):
			if m.overlay == nil {
				m.help.ShowAll = !m.help.ShowAll
				return m.resizeScreens()
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

func (m Model) resizeScreens() (tea.Model, tea.Cmd) {
	width, height := m.width, m.height

	m.help.SetWidth(width)

	height -= lipgloss.Height(m.renderHeader())
	height -= lipgloss.Height(m.renderFooter())

	msg := tea.WindowSizeMsg{
		Width:  width,
		Height: height,
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	for i := range m.tabs {
		m.tabs[i], cmd = m.tabs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.overlay != nil {
		m.overlay, cmd = m.overlay.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

var (
	logoStyle        = lipgloss.NewStyle().Bold(true).Padding(0, 1).Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	bulletStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

func (m Model) View() tea.View {
	header := m.renderHeader()
	footer := m.renderFooter()

	var content string
	if m.overlay != nil {
		content = m.overlay.View()
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			m.tabs[m.activeTab].View(),
			footer,
		)
	}

	content = appStyle.Render(content)

	v := tea.NewView(content)
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

	return logo + "\n\n" + tabBar + "\n"
}

func (m Model) renderFooter() string {
	var screenKM help.KeyMap
	if m.overlay != nil {
		screenKM = m.overlay.KeyMap()
	} else {
		screenKM = m.tabs[m.activeTab].KeyMap()
	}
	return "\n" + m.help.View(mergedKeyMap{
		screen:  screenKM,
		overlay: m.overlay != nil,
	})
}

// mergedKeyMap combines a screen's key map with app-level bindings.
type mergedKeyMap struct {
	screen  help.KeyMap
	overlay bool
}

func (k mergedKeyMap) ShortHelp() []key.Binding {
	keys := k.screen.ShortHelp()
	if !k.overlay {
		keys = append(keys, keyTab, keyToggleHelp)
	}
	return append(keys, keyQuit)
}

func (k mergedKeyMap) FullHelp() [][]key.Binding {
	groups := k.screen.FullHelp()
	var appKeys []key.Binding
	if !k.overlay {
		appKeys = append(appKeys, keyTab, keyShiftTab, keyToggleHelp)
	}
	return append(groups, append(appKeys, keyQuit))
}
