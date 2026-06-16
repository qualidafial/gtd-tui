package projectpicker

import (
	"context"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/selectfield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

type Model struct {
	task       gtd.Task
	taskSvc    gtd.TaskService
	projectSvc gtd.ProjectService
	original   int64
	form       form.Model
	loaded     bool
	saving     bool
}

func New(task gtd.Task, taskSvc gtd.TaskService, projectSvc gtd.ProjectService) Model {
	var projectID int64
	if task.ProjectID != nil {
		projectID = *task.ProjectID
	}

	// Build the form immediately with an empty option set; the project list
	// loads asynchronously (see loadCmd) and is populated in place via
	// form.UpdateField when projectsLoadedMsg arrives. WithInitialValue
	// re-applies once the real options land.
	fieldOpts := []selectfield.FieldOption[int64]{
		selectfield.WithNone[int64]("(none)"),
		selectfield.WithSubmitOnEnter[int64](),
	}
	if task.ProjectID != nil {
		fieldOpts = append(fieldOpts, selectfield.WithInitialValue(*task.ProjectID))
	}

	return Model{
		task:       task,
		taskSvc:    taskSvc,
		projectSvc: projectSvc,
		original:   projectID,
		form:       form.New(selectfield.New[int64]("project", "Project", nil, fieldOpts...)),
	}
}

func (m Model) Init() tea.Cmd { return tea.Batch(m.form.Init(), m.loadCmd()) }

func (m Model) loadCmd() tea.Cmd {
	svc := m.projectSvc
	return func() tea.Msg {
		projects, err := svc.ListProjects(context.Background(), gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusOpen))
		if err != nil {
			return fmt.Errorf("load projects: %w", err)
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case projectsLoadedMsg:
		opts := make([]selectfield.Option[int64], 0, len(msg.projects))
		for _, p := range msg.projects {
			opts = append(opts, selectfield.Option[int64]{Display: p.Title, Value: p.ID})
		}
		m.form = m.form.UpdateField("project", func(f form.Field) form.Field {
			return f.(selectfield.Model[int64]).SetOptions(opts)
		})
		m.loaded = true
		return m, nil
	case savedMsg:
		if msg.err != nil {
			m.saving = false
			err := msg.err
			return m, cmds.Emit(fmt.Errorf("save failed: %w", err))
		}
		return screen.Dismiss()
	}

	if m.saving {
		return m, nil
	}

	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	// Submission is only meaningful once the project list has loaded; before
	// then the field holds only the synthetic "(none)" entry.
	if _, ok := msg.(form.SubmittedMsg); ok && m.loaded {
		selected, _ := m.form.FieldValues()["project"].(int64)
		if selected == m.original {
			return screen.Dismiss()
		}
		m.saving = true
		return m, m.saveCmd(selected)
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) saveCmd(selected int64) tea.Cmd {
	task := m.task
	if selected == 0 {
		task.ProjectID = nil
	} else {
		task.ProjectID = new(selected)
	}
	svc := m.taskSvc
	return func() tea.Msg {
		_, err := svc.UpdateTask(context.Background(), task)
		if err != nil {
			slog.Error("assign project: " + err.Error())
		}
		return savedMsg{err: err}
	}
}

func (m Model) View() string {
	if !m.loaded {
		return "Loading projects..."
	}
	return m.form.View()
}

func (m Model) CapturingInput() bool { return !m.saving }

// Keys aggregates the form's resolved bindings plus this screen's own esc
// binding (Resolve subtracts the overlay's duplicate esc). esc works even
// while the project list is still loading.
func (m Model) Keys() []keymap.Group {
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}

type projectsLoadedMsg struct {
	projects []gtd.Project
}

type savedMsg struct {
	err error
}
