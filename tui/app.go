package tui

import (
	"context"
	"log/slog"
	"slices"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/tabcontainer"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox"
	"github.com/qualidafial/gtd-tui/tui/pages/projects"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectpicker"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectview"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskconvert"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
)

var (
	appStyle    = lipgloss.NewStyle().Margin(1, 2)
	appErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)

type Model struct {
	active screen.Screen
	help   help.Model
	KeyMap KeyMap
	err    error
	width  int
	height int
}

func New(
	taskSvc gtd.TaskService,
	projectSvc gtd.ProjectService,
	inboxSvc gtd.InboxService,
) Model {
	pickerFn := func(task gtd.Task) screen.Screen {
		return projectpicker.New(task, taskSvc, projectSvc)
	}
	convertFn := func(task gtd.Task) screen.Screen {
		projectViewFn := func(p gtd.Project) screen.Screen {
			return projectview.New(p, taskSvc, projectSvc, pickerFn)
		}
		return taskconvert.New(task, projectSvc, projectViewFn)
	}
	projectNameFn := func(id int64) string {
		p, err := projectSvc.GetProject(context.Background(), id)
		if err != nil {
			slog.Error("resolving project name", "id", id, "err", err)
			return ""
		}
		return p.Title
	}
	pending := tasklist.New(taskSvc, "status:open ready:now", pickerFn, convertFn, projectNameFn, true)
	projectList := projects.New(projectSvc, taskSvc, pickerFn)
	inboxPage := inbox.New(inboxSvc, taskSvc, projectSvc)

	tabs := tabcontainer.New(
		tabcontainer.Tab{Label: "Inbox", Screen: inboxPage},
		tabcontainer.Tab{Label: "Tasks", Screen: pending},
		tabcontainer.Tab{Label: "Projects", Screen: projectList},
	)

	return Model{
		active: tabs,
		help:   help.New(),
		KeyMap: DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, tea.RequestWindowSize, m.active.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case error:
		m.err = msg
		return m.resizeActive()
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
		// Any keypress clears the error bar; the key is still forwarded below.
		if m.err != nil {
			m.err = nil
			var cmd tea.Cmd
			m, cmd = m.resizeActive()
			cmds = append(cmds, cmd)
		}
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.Help):
			if !screen.CapturingInput(m.active) {
				m.help.ShowAll = !m.help.ShowAll
				return m.resizeActive()
			}
		}
	}

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	m.updateKeyBindings()
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateKeyBindings() {
	capturingInput := screen.CapturingInput(m.active)
	m.KeyMap.Help.SetEnabled(!capturingInput)
}

func (m Model) resizeActive() (Model, tea.Cmd) {
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
	if m.err != nil {
		return "\n" + appErrStyle.Render(m.err.Error())
	}
	return "\n" + m.help.View(m)
}

// Keys aggregates the active screen's subtree (highest priority) ahead
// of the app's global bindings, so a key claimed by the active screen wins
// and is subtracted from the app's help.
func (m Model) Keys() []keymap.Group {
	return slices.Concat(
		m.active.Keys(),
		m.KeyMap.Keys(),
	)
}

// ShortHelp / FullHelp render the footer via the help component. Both are
// projections of a single Resolve pass over the aggregated bindings, so a
// key claimed by a higher-priority layer never double-lists.
func (m Model) ShortHelp() []key.Binding {
	return keymap.ShortHelp(m.Keys())
}

func (m Model) FullHelp() [][]key.Binding {
	return keymap.FullHelp(m.Keys())
}
