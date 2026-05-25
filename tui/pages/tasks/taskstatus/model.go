package taskstatus

import (
	"context"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/date"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks"
)

type Model struct {
	task       gtd.Task
	svc        gtd.TaskService
	transition Transition
	// confirm and at are pointers so the bound huh fields and every Model copy
	// that flows through the screen stack share one stable address; binding to
	// a value field writes to a stale copy and the input is lost. at is a
	// **time.Time because the date field reassigns the *time.Time it edits
	// (to nil when cleared, or to a fresh parsed time), so the shared storage
	// must be a stable slot holding that pointer, not the pointer itself.
	confirm  *bool
	at       **time.Time
	err      error
	form     *huh.Form
	applying bool
}

func New(task gtd.Task, svc gtd.TaskService, transition Transition) Model {
	m := Model{
		task:       task,
		svc:        svc,
		transition: transition,
		confirm:    new(true),            // default selection is the affirmative button
		at:         new(new(time.Now())), // transition timestamp defaults to now
	}

	s := specs[transition]
	when := date.NewField().
		Title("When").
		Value(m.at)
	confirm := huh.NewConfirm().
		Title(s.title).
		Description(s.description(task.Title)).
		Affirmative(s.affirmative).
		Negative("Cancel").
		Value(m.confirm)

	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

	m.form = huh.NewForm(huh.NewGroup(when, confirm)).
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
	// A cleared timestamp field leaves the slot nil; fall back to now so the
	// instant is never null and the Enter-Enter path needs no special value.
	at := time.Now()
	if m.at != nil && *m.at != nil {
		at = **m.at
	}
	return func() tea.Msg {
		_, err := apply(svc, context.Background(), id, at)
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
