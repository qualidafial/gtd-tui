package taskedit

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
	metaLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(8)
	metaValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type Model struct {
	task        gtd.Task
	svc         gtd.TaskService
	projectName string
	err         error
	form        form.Model
	saving      bool
}

func New(task gtd.Task, svc gtd.TaskService, projectName string) Model {
	if task.ID == 0 {
		task.Status = gtd.TaskStatusOpen
	}

	assignee := ""
	if task.Assignee != nil {
		assignee = *task.Assignee
	}

	title := inputfield.New("title", "Title",
		inputfield.WithValue(task.Title),
		inputfield.WithValidator(func(s string) error {
			if len(s) == 0 {
				return errors.New("title is required")
			}
			return nil
		}),
	)
	desc := textfield.New("description", "Description",
		textfield.WithValue(task.Description),
	)
	asg := inputfield.New("assignee", "Assignee",
		inputfield.WithValue(assignee),
	)

	dueOpts := []datefield.Option{}
	if task.Due != nil {
		dueOpts = append(dueOpts, datefield.WithValue(task.Due))
	}
	due := datefield.New("due", "Due", dueOpts...)

	deferOpts := []datefield.Option{}
	if task.DeferUntil != nil {
		deferOpts = append(deferOpts, datefield.WithValue(task.DeferUntil))
	}
	deferUntil := datefield.New("defer", "Defer Until", deferOpts...)

	saveLabel := "Save"
	if task.ID == 0 {
		saveLabel = "Create"
	}
	save := savefield.New("save", savefield.WithLabel(saveLabel))

	return Model{
		task:        task,
		svc:         svc,
		projectName: projectName,
		form:        form.New(title, desc, asg, due, deferUntil, save),
	}
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
			m.err = nil
		}
		return m, nil
	}

	if m.saving {
		if sm, ok := msg.(taskSavedMsg); ok {
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
	case taskSavedMsg:
		return m.handleSaved(msg)
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) handleSaved(msg taskSavedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("save failed: %w", err))
	}
	return screen.Dismiss()
}

func (m Model) saveCmd() tea.Cmd {
	values := m.form.FieldValues()
	task := m.task
	task.Title, _ = values["title"].(string)
	task.Description, _ = values["description"].(string)
	if asg, _ := values["assignee"].(string); asg != "" {
		task.Assignee = new(asg)
	} else {
		task.Assignee = nil
	}
	task.Due, _ = values["due"].(*time.Time)
	task.DeferUntil, _ = values["defer"].(*time.Time)

	svc := m.svc
	return func() tea.Msg {
		var saved gtd.Task
		var err error
		ctx := context.Background()
		if task.ID == 0 {
			saved, err = svc.CreateTask(ctx, task)
		} else {
			saved, err = svc.UpdateTask(ctx, task)
		}
		if err != nil {
			slog.Error("saving task: " + err.Error())
		}
		return taskSavedMsg{task: saved, err: err}
	}
}

func (m Model) View() string {
	var sections []string
	if m.task.ID != 0 {
		sections = append(sections,
			m.metaLine("Task ID", fmt.Sprint(m.task.ID)),
			m.metaLine("Created", m.task.CreatedAt.Local().Format(time.DateTime)),
			m.metaLine("Updated", m.task.UpdatedAt.Local().Format(time.DateTime)),
			m.metaLine("Status", m.statusValue()),
		)
		if m.projectName != "" {
			sections = append(sections, m.metaLine("Project", m.projectName))
		}
		sections = append(sections, "")
	}
	sections = append(sections, m.form.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) metaLine(label, value string) string {
	return metaLabelStyle.Render(label+":") + " " + metaValueStyle.Render(value)
}

func (m Model) statusValue() string {
	when := reltime.Format(m.task.StatusChangedAt, time.Now())
	return fmt.Sprintf("%s (%s)", titleStatus(m.task.Status), when)
}

func titleStatus(s gtd.TaskStatus) string {
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

type taskSavedMsg struct {
	task gtd.Task
	err  error
}
