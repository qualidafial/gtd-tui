package projectedit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/reltime"
	"github.com/qualidafial/gtd-tui/tui/components/date"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

var (
	metaLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(11)
	metaValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type ViewFactory func(project gtd.Project) screen.Screen

type Model struct {
	project     *gtd.Project
	svc         gtd.ProjectService
	viewFactory ViewFactory
	err         error
	form        *huh.Form
	saving      bool
}

func New(project gtd.Project, svc gtd.ProjectService, viewFactory ViewFactory) Model {
	if project.ID == 0 {
		project.Status = gtd.ProjectStatusOpen
	}

	m := Model{
		project:     &project,
		svc:         svc,
		viewFactory: viewFactory,
	}

	fields := []huh.Field{
		huh.NewInput().
			Title("Title").
			Value(&project.Title).
			Validate(func(s string) error {
				if len(s) == 0 {
					return errors.New("title is required")
				}
				return nil
			}),
		huh.NewInput().
			Title("Outcome").
			Value(&project.Outcome).
			Validate(func(s string) error {
				if len(s) == 0 {
					return errors.New("outcome is required")
				}
				return nil
			}),
		huh.NewText().
			Title("Description").
			Value(&project.Description),
		date.NewField().
			Title("Due").
			Value(&project.Due),
	}
	group := huh.NewGroup(fields...)

	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = keyBack

	m.form = huh.NewForm(group).
		WithShowErrors(true).
		WithShowHelp(false).
		WithKeyMap(keymap)
	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case projectSavedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.saving = false
			err := msg.err
			return m, func() tea.Msg { return fmt.Errorf("save failed: %w", err) }
		}
		if msg.created && m.viewFactory != nil {
			return screen.Replace(m.viewFactory(msg.project))
		}
		return screen.Dismiss()
	case tea.KeyPressMsg:
		if m.err != nil {
			if key.Matches(msg, keyBack) {
				m.err = nil
				m.form.State = huh.StateNormal
			}
			return m, nil
		}
	}

	if m.saving {
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	switch m.form.State {
	case huh.StateAborted:
		return screen.Dismiss(cmd)
	case huh.StateCompleted:
		m.saving = true
		return m, tea.Batch(cmd, m.saveCmd())
	}
	return m, cmd
}

func (m Model) saveCmd() tea.Cmd {
	project := *m.project
	svc := m.svc
	creating := project.ID == 0
	return func() tea.Msg {
		var saved gtd.Project
		var err error
		ctx := context.Background()
		if creating {
			saved, err = svc.CreateProject(ctx, project)
		} else {
			saved, err = svc.UpdateProject(ctx, project)
		}
		if err != nil {
			slog.Error("saving project: " + err.Error())
		}
		return projectSavedMsg{
			project: saved,
			created: creating,
			err:     err,
		}
	}
}

func (m Model) View() string {
	var sections []string
	if m.project.ID != 0 {
		sections = append(sections,
			m.metaLine("Project ID", fmt.Sprint(m.project.ID)),
			m.metaLine("Status", m.statusValue()),
			m.metaLine("Created", m.project.CreatedAt.Local().Format(time.DateTime)),
			m.metaLine("Updated", m.project.UpdatedAt.Local().Format(time.DateTime)),
			"",
		)
	}
	sections = append(sections, m.form.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) metaLine(label, value string) string {
	return metaLabelStyle.Render(label+":") + " " + metaValueStyle.Render(value)
}

func (m Model) statusValue() string {
	when := reltime.Format(m.project.StatusChangedAt, time.Now())
	return fmt.Sprintf("%s (%s)", titleStatus(m.project.Status), when)
}

func titleStatus(s gtd.ProjectStatus) string {
	str := string(s)
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

func (m Model) CapturingInput() bool {
	return m.form.State == huh.StateNormal
}

func (m Model) ShortHelp() []key.Binding {
	return m.form.KeyBinds()
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.form.KeyBinds()}
}

type projectSavedMsg struct {
	project gtd.Project
	created bool
	err     error
}
