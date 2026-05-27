package projectcreate

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

var (
	keySubmit = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "create"))
	keyCancel = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
)

type Model struct {
	svc      gtd.ProjectService
	input    textinput.Model
	creating bool
	err      error
}

func New(svc gtd.ProjectService) Model {
	ti := textinput.New()
	ti.Placeholder = "Project title"
	m := Model{
		svc:   svc,
		input: ti,
		err:   nil,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return initCmd
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case initMsg:
		cmd := m.input.Focus()
		return m, cmd
	case tea.WindowSizeMsg:
		m.input.SetWidth(msg.Width)
		return m, nil
	case error:
		m.creating = false
		m.err = msg
		return m, nil
	case tea.KeyPressMsg:
		m.err = nil
		switch {
		case key.Matches(msg, keyCancel):
			return m, screen.Dismiss()
		case key.Matches(msg, keySubmit):
			if !m.creating {
				title := strings.TrimSpace(m.input.Value())
				if title == "" {
					m.err = fmt.Errorf("title cannot be empty")
					return m, nil
				}
				m.creating = true
				return m, m.createCmd(title)
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) createCmd(title string) tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		_, err := svc.CreateProject(context.Background(), gtd.Project{
			Title:  title,
			Status: gtd.ProjectStatusOpen,
		})
		if err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		return screen.DismissMsg{}
	}
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("New Project"))
	b.WriteByte('\n')
	b.WriteString(m.input.View())
	b.WriteByte('\n')
	if m.err != nil {
		b.WriteString(errStyle.Render(m.err.Error()))
	}
	return b.String()
}

func (m Model) CapturingInput() bool { return true }

func (m Model) KeyMap() help.KeyMap { return keymap{} }

func initCmd() tea.Msg {
	return initMsg{}
}

type initMsg struct{}

type keymap struct{}

func (keymap) ShortHelp() []key.Binding  { return []key.Binding{keySubmit, keyCancel} }
func (keymap) FullHelp() [][]key.Binding { return [][]key.Binding{{keySubmit, keyCancel}} }
