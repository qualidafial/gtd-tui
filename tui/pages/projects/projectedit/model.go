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
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/reltime"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/datefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/textfield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

var (
	metaLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(11)
	metaValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type ViewFactory func(project gtd.Project) screen.Screen

type Model struct {
	project     gtd.Project
	svc         gtd.ProjectService
	viewFactory ViewFactory
	err         error
	form        form.Model
	saving      bool
}

func New(project gtd.Project, svc gtd.ProjectService, viewFactory ViewFactory) Model {
	if project.ID == 0 {
		project.Status = gtd.ProjectStatusOpen
	}

	requireNonEmpty := func(label string) func(string) error {
		return func(s string) error {
			if len(s) == 0 {
				return errors.New(label + " is required")
			}
			return nil
		}
	}

	title := inputfield.New("title", "Title",
		inputfield.WithValue(project.Title),
		inputfield.WithValidator(requireNonEmpty("title")),
	)
	outcome := inputfield.New("outcome", "Outcome",
		inputfield.WithValue(project.Outcome),
		inputfield.WithValidator(requireNonEmpty("outcome")),
	)
	desc := textfield.New("description", "Description",
		textfield.WithValue(project.Description),
	)
	dueOpts := []datefield.Option{}
	if project.Due != nil {
		dueOpts = append(dueOpts, datefield.WithValue(project.Due))
	}
	due := datefield.New("due", "Due", dueOpts...)

	saveLabel := "Save"
	if project.ID == 0 {
		saveLabel = "Create"
	}
	save := savefield.New("save", savefield.WithLabel(saveLabel))

	return Model{
		project:     project,
		svc:         svc,
		viewFactory: viewFactory,
		form:        form.New(title, outcome, desc, due, save),
	}
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	// Save-error standoff: an earlier save failed; Esc clears the error.
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
			m.err = nil
		}
		return m, nil
	}

	if m.saving {
		if sm, ok := msg.(projectSavedMsg); ok {
			return m.handleSaved(sm)
		}
		return m, nil
	}

	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	switch msg := msg.(type) {
	case form.SubmittedMsg:
		_ = msg
		m.saving = true
		return m, m.saveCmd()
	case projectSavedMsg:
		return m.handleSaved(msg)
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) handleSaved(msg projectSavedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("save failed: %w", err))
	}
	if msg.created && m.viewFactory != nil {
		return screen.Replace(m.viewFactory(msg.project))
	}
	return screen.Dismiss()
}

func (m Model) saveCmd() tea.Cmd {
	project := m.project
	values := m.form.FieldValues()
	project.Title, _ = values["title"].(string)
	project.Outcome, _ = values["outcome"].(string)
	project.Description, _ = values["description"].(string)
	project.Due, _ = values["due"].(*time.Time)

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

func (m Model) CapturingInput() bool { return m.err == nil && !m.saving }

func (m Model) ShortHelp() []key.Binding {
	return append(m.form.ShortHelp(), keyBack)
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{append(m.form.ShortHelp(), keyBack)}
}

type projectSavedMsg struct {
	project gtd.Project
	created bool
	err     error
}
