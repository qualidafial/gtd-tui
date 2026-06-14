// Package taskpicker is a selection-only overlay for choosing a standalone
// open task. On confirm it broadcasts the chosen task to the calling screen
// via SelectedMsg and dismisses; it applies no change itself. The calling
// screen (the project view) owns linking the task and reporting any error.
//
// The caller is responsible for opening the picker only when at least one
// standalone open task exists, so the picker renders no empty state.
package taskpicker

import (
	"context"
	"fmt"

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

// SelectedMsg carries the task the user chose. The calling screen handles it
// (e.g. by linking the task into a project) once it is active again.
type SelectedMsg struct {
	Task gtd.Task
}

type Model struct {
	taskSvc gtd.TaskService
	tasks   []gtd.Task
	form    form.Model
	ready   bool
}

func New(taskSvc gtd.TaskService) Model {
	return Model{taskSvc: taskSvc}
}

func (m Model) Init() tea.Cmd { return m.loadCmd() }

func (m Model) loadCmd() tea.Cmd {
	svc := m.taskSvc
	return func() tea.Msg {
		tasks, err := svc.ListTasks(context.Background(), gtd.TaskFilter{Status: new(gtd.TaskStatusOpen)})
		if err != nil {
			return fmt.Errorf("load tasks: %w", err)
		}
		// Selection-only candidates are standalone open tasks.
		standalone := tasks[:0:0]
		for _, t := range tasks {
			if t.ProjectID == nil {
				standalone = append(standalone, t)
			}
		}
		return tasksLoadedMsg{tasks: standalone}
	}
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.form = m.buildForm(msg.tasks)
		m.ready = true
		return m, m.form.Init()
	}

	if !m.ready {
		return m, nil
	}

	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	if _, ok := msg.(form.SubmittedMsg); ok {
		selected, _ := m.form.FieldValues()["task"].(int64)
		for _, t := range m.tasks {
			if t.ID == selected {
				return screen.Dismiss(cmds.Emit(SelectedMsg{Task: t}))
			}
		}
		// No matching task (empty candidate set should not reach here); dismiss
		// without broadcasting.
		return screen.Dismiss()
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) buildForm(tasks []gtd.Task) form.Model {
	opts := make([]selectfield.Option[int64], 0, len(tasks))
	for _, t := range tasks {
		opts = append(opts, selectfield.Option[int64]{Display: t.Title, Value: t.ID})
	}
	return form.New(selectfield.New("task", "Task", opts,
		selectfield.WithSubmitOnEnter[int64](),
	))
}

func (m Model) View() string {
	if !m.ready {
		return "Loading tasks..."
	}
	return m.form.View()
}

func (m Model) CapturingInput() bool { return m.ready }

// Keys returns nothing until the task list has loaded; afterward it aggregates
// the form's resolved bindings plus this screen's own esc binding (Resolve
// subtracts the overlay's duplicate esc).
func (m Model) Keys() []keymap.Group {
	if !m.ready {
		return nil
	}
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}

type tasksLoadedMsg struct {
	tasks []gtd.Task
}
