// Package taskconvert is the convert-to-project wizard. It promotes a
// standalone task into a new open project, keeping the task as the project's
// first action. The flow is form-first: it collects the new project's
// Title/Outcome/Description and the re-scoped task's Title/Description, then
// commits via ProjectService.ConvertTaskToProject in a single transaction on
// submit. Because the source task already exists and is durable, there is no
// early checkpoint — abandoning the wizard (esc) leaves the task unchanged and
// standalone.
package taskconvert

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/textfield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

type Model struct {
	task   gtd.Task
	svc    gtd.ProjectService
	form   form.Model
	err    error
	saving bool
}

func New(task gtd.Task, svc gtd.ProjectService) Model {
	requireNonEmpty := func(label string) func(string) error {
		return func(s string) error {
			if len(s) == 0 {
				return errors.New(label + " is required")
			}
			return nil
		}
	}

	projectTitle := inputfield.New("project_title", "Project title",
		inputfield.WithValue(task.Title),
		inputfield.WithValidator(requireNonEmpty("project title")),
	)
	outcome := inputfield.New("outcome", "Outcome",
		inputfield.WithValidator(requireNonEmpty("outcome")),
	)
	projectDesc := textfield.New("project_description", "Project description",
		textfield.WithValue(task.Description),
	)
	taskTitle := inputfield.New("task_title", "First action",
		inputfield.WithValue(task.Title),
		inputfield.WithValidator(requireNonEmpty("first action")),
	)
	taskDesc := textfield.New("task_description", "Action notes",
		textfield.WithValue(task.Description),
	)
	save := savefield.New("save", savefield.WithLabel("Convert"))

	return Model{
		task: task,
		svc:  svc,
		form: form.New(projectTitle, outcome, projectDesc, taskTitle, taskDesc, save),
	}
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	// Save-error standoff: an earlier commit failed; Esc clears the error.
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
			m.err = nil
		}
		return m, nil
	}

	if m.saving {
		if cm, ok := msg.(convertedMsg); ok {
			return m.handleConverted(cm)
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
		return m, m.convertCmd()
	case convertedMsg:
		return m.handleConverted(msg)
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) handleConverted(msg convertedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("convert failed: %w", err))
	}
	return screen.Dismiss()
}

func (m Model) convertCmd() tea.Cmd {
	values := m.form.FieldValues()
	project := gtd.Project{}
	project.Title, _ = values["project_title"].(string)
	project.Outcome, _ = values["outcome"].(string)
	project.Description, _ = values["project_description"].(string)

	reframed := gtd.Task{}
	reframed.Title, _ = values["task_title"].(string)
	reframed.Description, _ = values["task_description"].(string)

	svc := m.svc
	taskID := m.task.ID
	return func() tea.Msg {
		_, _, err := svc.ConvertTaskToProject(context.Background(), taskID, project, reframed)
		if err != nil {
			slog.Error("converting task to project: " + err.Error())
		}
		return convertedMsg{err: err}
	}
}

func (m Model) View() string { return m.form.View() }

func (m Model) CapturingInput() bool { return m.err == nil && !m.saving }

// Keys aggregates the form's resolved bindings and appends this screen's own
// esc binding as a trailing group; Resolve subtracts the overlay's duplicate
// esc.
func (m Model) Keys() []keymap.Group {
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}

type convertedMsg struct {
	err error
}
