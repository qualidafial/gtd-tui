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
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

type Model struct {
	task       gtd.Task
	taskSvc    gtd.TaskService
	projectSvc gtd.ProjectService
	original   int64
	form       form.Model
	ready      bool
	saving     bool
}

func New(task gtd.Task, taskSvc gtd.TaskService, projectSvc gtd.ProjectService) Model {
	var projectID int64
	if task.ProjectID != nil {
		projectID = *task.ProjectID
	}
	return Model{
		task:       task,
		taskSvc:    taskSvc,
		projectSvc: projectSvc,
		original:   projectID,
	}
}

func (m Model) Init() tea.Cmd { return m.loadCmd() }

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
		m.form = m.buildForm(msg.projects)
		m.ready = true
		return m, m.form.Init()
	case savedMsg:
		if msg.err != nil {
			m.saving = false
			err := msg.err
			return m, cmds.Emit(fmt.Errorf("save failed: %w", err))
		}
		return screen.Dismiss()
	}

	if !m.ready || m.saving {
		return m, nil
	}

	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	if _, ok := msg.(form.SubmittedMsg); ok {
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

func (m Model) buildForm(projects []gtd.Project) form.Model {
	opts := make([]selectfield.Option[int64], 0, len(projects))
	for _, p := range projects {
		opts = append(opts, selectfield.Option[int64]{Display: p.Title, Value: p.ID})
	}

	// Initial selection (when the task already has a project) is applied
	// at construction via WithInitialValue so the index settles after
	// WithNone has prepended its synthetic option.
	if m.task.ProjectID != nil {
		return form.New(selectfield.New("project", "Project", opts,
			selectfield.WithNone[int64]("(none)"),
			selectfield.WithSubmitOnEnter[int64](),
			selectfield.WithInitialValue(*m.task.ProjectID),
		))
	}
	return form.New(selectfield.New("project", "Project", opts,
		selectfield.WithNone[int64]("(none)"),
		selectfield.WithSubmitOnEnter[int64](),
	))
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
	if !m.ready {
		return "Loading projects..."
	}
	return m.form.View()
}

func (m Model) CapturingInput() bool { return m.ready && !m.saving }

func (m Model) ShortHelp() []key.Binding {
	if !m.ready {
		return nil
	}
	return append(m.form.ShortHelp(), keyBack)
}

func (m Model) FullHelp() [][]key.Binding {
	if !m.ready {
		return nil
	}
	return [][]key.Binding{append(m.form.ShortHelp(), keyBack)}
}

type projectsLoadedMsg struct {
	projects []gtd.Project
}

type savedMsg struct {
	err error
}
