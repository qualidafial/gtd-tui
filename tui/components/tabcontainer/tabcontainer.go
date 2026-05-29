package tabcontainer

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

var (
	keyTab      = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next view"))
	keyShiftTab = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev view"))
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
	activeTab int
	width     int
	height    int
}

func New(tabs ...Tab) Model {
	return Model{tabs: tabs}
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
		return m, cmd
	case tea.KeyPressMsg:
		if !screen.CapturingInput(m.tabs[m.activeTab].Screen) {
			switch {
			case key.Matches(msg, keyTab):
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				return m, m.tabs[m.activeTab].Screen.Init()
			case key.Matches(msg, keyShiftTab):
				m.activeTab = (m.activeTab + len(m.tabs) - 1) % len(m.tabs)
				return m, m.tabs[m.activeTab].Screen.Init()
			}
		}
	}

	var cmd tea.Cmd
	m.tabs[m.activeTab].Screen, cmd = m.tabs[m.activeTab].Screen.Update(msg)
	return m, cmd
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

func (m Model) KeyMap() help.KeyMap {
	return tabKeyMap{
		inner:          m.tabs[m.activeTab].Screen.KeyMap(),
		capturingInput: screen.CapturingInput(m.tabs[m.activeTab].Screen),
	}
}

func (m Model) CapturingInput() bool {
	return screen.CapturingInput(m.tabs[m.activeTab].Screen)
}

type tabKeyMap struct {
	inner          help.KeyMap
	capturingInput bool
}

func (k tabKeyMap) ShortHelp() []key.Binding {
	if k.capturingInput {
		return k.inner.ShortHelp()
	}
	return append(k.inner.ShortHelp(), keyTab)
}

func (k tabKeyMap) FullHelp() [][]key.Binding {
	groups := k.inner.FullHelp()
	if k.capturingInput {
		return groups
	}
	return append(groups, []key.Binding{keyTab, keyShiftTab})
}