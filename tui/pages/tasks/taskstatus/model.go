package taskstatus

import (
	"context"
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
	task       gtd.Task
	svc        gtd.TaskService
	transition Transition
	// confirm is a pointer so the bound huh field and every Model copy that
	// flows through the screen stack share one stable address; binding to a
	// value field writes to a stale copy and the confirmation is lost.
	confirm  *bool
	err      error
	form     *huh.Form
	applying bool
}

func New(task gtd.Task, svc gtd.TaskService, transition Transition) Model {
	m := Model{
		task:       task,
		svc:        svc,
		transition: transition,
		confirm:    new(true), // default selection is the affirmative button
	}

	s := specs[transition]
	field := huh.NewConfirm().
		Title(s.title).
		Description(s.description(task.Title)).
		Affirmative(s.affirmative).
		Negative("Cancel").
		Value(m.confirm)

	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

	m.form = huh.NewForm(huh.NewGroup(field)).
		WithShowErrors(true).
		WithShowHelp(false).
		WithKeyMap(keymap)
	return m
}

// Transition reports which status change this overlay will apply on confirm.
func (m Model) Transition() Transition {
	return m.transition
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case taskTransitionedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.applying = false
			return m, nil
		}
		return m, tea.Batch(screen.HideOverlay(), tasks.TasksChanged())
	}

	if m.applying {
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
		if !*m.confirm {
			return m, tea.Batch(cmd, screen.HideOverlay())
		}
		m.applying = true
		return m, tea.Batch(cmd, m.applyCmd())
	}
	return m, cmd
}

func (m Model) applyCmd() tea.Cmd {
	id := m.task.ID
	svc := m.svc
	apply := specs[m.transition].apply
	return func() tea.Msg {
		_, err := apply(svc, context.Background(), id)
		if err != nil {
			slog.Error("transitioning task: " + err.Error())
		}
		return taskTransitionedMsg{err: err}
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

type taskTransitionedMsg struct {
	err error
}
