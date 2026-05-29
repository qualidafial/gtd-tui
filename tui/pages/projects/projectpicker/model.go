package projectpicker

import (
	"context"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

type Model struct {
	task       gtd.Task
	taskSvc    gtd.TaskService
	projectSvc gtd.ProjectService
	selected   **int64
	original   **int64
	form       *huh.Form
	saving     bool
}

func New(task gtd.Task, taskSvc gtd.TaskService, projectSvc gtd.ProjectService) Model {
	m := Model{
		task:       task,
		taskSvc:    taskSvc,
		projectSvc: projectSvc,
		selected:   new(task.ProjectID),
		original:   new(task.ProjectID),
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return m.loadCmd()
}

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
		m.buildForm(msg.projects)
		return m, m.form.Init()
	case savedMsg:
		if msg.err != nil {
			m.saving = false
			if m.form != nil {
				m.form.State = huh.StateNormal
			}
			err := msg.err
			return m, func() tea.Msg { return fmt.Errorf("save failed: %w", err) }
		}
		return m, screen.Dismiss()
	}

	if m.form == nil || m.saving {
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	switch m.form.State {
	case huh.StateAborted:
		return m, tea.Batch(cmd, screen.Dismiss())
	case huh.StateCompleted:
		if ptrEqual(*m.selected, *m.original) {
			return m, tea.Batch(cmd, screen.Dismiss())
		}
		m.saving = true
		return m, tea.Batch(cmd, m.saveCmd())
	}
	return m, cmd
}

func (m *Model) buildForm(projects []gtd.Project) {
	options := make([]huh.Option[*int64], 0, len(projects)+1)
	options = append(options, huh.NewOption("(none)", (*int64)(nil)))
	for _, p := range projects {
		id := new(p.ID)
		options = append(options, huh.NewOption(p.Title, id))
		if m.task.ProjectID != nil && *m.task.ProjectID == p.ID {
			*m.selected = id
			*m.original = new(*m.task.ProjectID)
		}
	}

	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = keyBack

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*int64]().
				Title("Project").
				Options(options...).
				Value(m.selected),
		),
	).
		WithShowErrors(true).
		WithShowHelp(false).
		WithKeyMap(keymap)
}

func (m Model) saveCmd() tea.Cmd {
	task := m.task
	task.ProjectID = *m.selected
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
	if m.form == nil {
		return "Loading projects..."
	}
	return m.form.View()
}

func (m Model) CapturingInput() bool {
	return m.form != nil && m.form.State == huh.StateNormal
}

func (m Model) KeyMap() help.KeyMap {
	if m.form == nil {
		return emptyKeyMap{}
	}
	return formKeyMap{m.form.KeyBinds()}
}

func ptrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

type projectsLoadedMsg struct {
	projects []gtd.Project
}

type savedMsg struct {
	err error
}

type formKeyMap struct {
	binds []key.Binding
}

func (k formKeyMap) ShortHelp() []key.Binding  { return k.binds }
func (k formKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.binds} }

type emptyKeyMap struct{}

func (emptyKeyMap) ShortHelp() []key.Binding  { return nil }
func (emptyKeyMap) FullHelp() [][]key.Binding { return nil }
