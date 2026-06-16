package taskstatus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/datefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/theme"
)

var (
	keyBack    = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	titleStyle = theme.Title
	descStyle  = theme.Subtitle
)

type Model struct {
	task       gtd.Task
	svc        gtd.TaskService
	transition Transition
	form       form.Model
	applying   bool
}

func New(task gtd.Task, svc gtd.TaskService, transition Transition) Model {
	s := specs[transition]
	now := time.Now()

	when := datefield.New("when", "When", datefield.WithValue(&now))
	save := savefield.New("save", savefield.WithLabel(s.affirmative))

	return Model{
		task:       task,
		svc:        svc,
		transition: transition,
		form:       form.New(when, save),
	}
}

// Transition reports which status change this overlay will apply on confirm.
func (m Model) Transition() Transition { return m.transition }

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if m.applying {
		if tm, ok := msg.(taskTransitionedMsg); ok {
			if tm.err != nil {
				m.applying = false
				err := tm.err
				return m, cmds.Emit(fmt.Errorf("transition failed: %w", err))
			}
			return screen.Dismiss()
		}
		return m, nil
	}

	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	switch msg := msg.(type) {
	case form.SubmittedMsg:
		_ = msg
		m.applying = true
		return m, m.applyCmd()
	case taskTransitionedMsg:
		if msg.err != nil {
			m.applying = false
			err := msg.err
			return m, cmds.Emit(fmt.Errorf("transition failed: %w", err))
		}
		return screen.Dismiss()
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

// applyCmd is exposed (lowercase but accessible within package tests) so a
// test that has driven the form to a valid state can invoke the apply step
// directly. The When value is read from the form; a cleared field falls back
// to now so the instant is never null.
func (m Model) applyCmd() tea.Cmd {
	id := m.task.ID
	svc := m.svc
	apply := specs[m.transition].apply
	at := time.Now()
	if t, _ := m.form.FieldValues()["when"].(*time.Time); t != nil {
		at = *t
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
	s := specs[m.transition]
	header := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(s.title),
		descStyle.Render(s.description(m.task.Title)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.form.View())
}

func (m Model) CapturingInput() bool { return !m.applying }

// Keys aggregates the form's resolved bindings and appends this screen's
// own esc binding as a trailing group; Resolve subtracts the overlay's
// duplicate esc.
func (m Model) Keys() []keymap.Group {
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}

type taskTransitionedMsg struct {
	err error
}
