package taskdelete

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
	"github.com/qualidafial/gtd-tui/tui/pages/tasks"
)

type Model struct {
	task     gtd.Task
	svc      gtd.TaskService
	confirm  bool
	err      error
	form     *huh.Form
	deleting bool
}

func New(task gtd.Task, svc gtd.TaskService) Model {
	m := Model{
		task: task,
		svc:  svc,
	}

	field := huh.NewConfirm().
		Title("Delete task?").
		Description(fmt.Sprintf("%q will be permanently deleted.", task.Title)).
		Affirmative("Delete").
		Negative("Cancel").
		Value(&m.confirm)

	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

	m.form = huh.NewForm(huh.NewGroup(field)).
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
	case taskDeletedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.deleting = false
			return m, nil
		}
		return m, tea.Batch(screen.HideOverlay(), tasks.TasksChanged())
	}

	if m.deleting {
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	switch m.form.State {
	case huh.StateAborted:
		return m, tea.Batch(cmd, screen.HideOverlay())
	case huh.StateCompleted:
		if !m.confirm {
			return m, tea.Batch(cmd, screen.HideOverlay())
		}
		m.deleting = true
		return m, tea.Batch(cmd, m.deleteCmd())
	}
	return m, cmd
}

func (m Model) deleteCmd() tea.Cmd {
	id := m.task.ID
	svc := m.svc
	return func() tea.Msg {
		err := svc.DeleteTask(context.Background(), id)
		if err != nil {
			slog.Error("deleting task: " + err.Error())
		}
		return taskDeletedMsg{err: err}
	}
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return m.form.View()
}

func (m Model) KeyMap() help.KeyMap {
	return KeyMap{m.form.KeyBinds()}
}

type taskDeletedMsg struct {
	err error
}
