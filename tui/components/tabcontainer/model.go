package tabcontainer

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

var (
	logoStyle        = lipgloss.NewStyle().Bold(true).Padding(0, 1).Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	bulletStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

type Tab struct {
	Label  string
	Screen screen.Screen
}

type Model struct {
	tabs      []Tab
	KeyMap    KeyMap
	activeTab int
	width     int
	height    int
}

func New(tabs ...Tab) Model {
	return Model{
		tabs:   tabs,
		KeyMap: DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.tabs[m.activeTab].Screen.Init()
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := lipgloss.Height(m.renderHeader())
		innerMsg := tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height - headerHeight,
		}
		var cmd tea.Cmd
		m.tabs[m.activeTab].Screen, cmd = m.tabs[m.activeTab].Screen.Update(innerMsg)
		m.updateKeyBindings()
		return m, cmd
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Next):
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			return m, m.tabs[m.activeTab].Screen.Init()
		case key.Matches(msg, m.KeyMap.Prev):
			m.activeTab = (m.activeTab + len(m.tabs) - 1) % len(m.tabs)
			return m, m.tabs[m.activeTab].Screen.Init()
		}
	}

	var cmd tea.Cmd
	m.tabs[m.activeTab].Screen, cmd = m.tabs[m.activeTab].Screen.Update(msg)
	m.updateKeyBindings()
	return m, cmd
}

func (m *Model) updateKeyBindings() {
	capturingInput := screen.CapturingInput(m.tabs[m.activeTab].Screen)
	m.KeyMap.Next.SetEnabled(!capturingInput)
	m.KeyMap.Prev.SetEnabled(!capturingInput)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.tabs[m.activeTab].Screen.View(),
	)
}

func (m Model) renderHeader() string {
	logo := logoStyle.Render("GTD")

	var tabs []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, bulletStyle.Render("•")+" "+activeTabStyle.Render(tab.Label))
		} else {
			tabs = append(tabs, "  "+inactiveTabStyle.Render(tab.Label))
		}
	}
	tabBar := strings.Join(tabs, "   ")

	return logo + "\n\n" + tabBar + "\n"
}

func (m Model) ShortHelp() []key.Binding {
	return slices.Concat(
		m.KeyMap.ShortHelp(),
		m.tabs[m.activeTab].Screen.ShortHelp(),
	)
}

func (m Model) FullHelp() [][]key.Binding {
	return slices.Concat(
		m.KeyMap.FullHelp(),
		m.tabs[m.activeTab].Screen.FullHelp(),
	)
}

func (m Model) CapturingInput() bool {
	return screen.CapturingInput(m.tabs[m.activeTab].Screen)
}
