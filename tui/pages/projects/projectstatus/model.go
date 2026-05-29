package projectstatus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/date"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

// Transition identifies a project status change initiated from the project list.
type Transition int

const (
	Complete Transition = iota
	Drop
)

type spec struct {
	title       string
	description func(title string, pending int) string
	affirmative string
	apply       func(svc gtd.ProjectService, ctx context.Context, id int64, at time.Time) (gtd.Project, error)
}

var specs = map[Transition]spec{
	Complete: {
		title: "Complete project?",
		description: func(t string, pending int) string {
			if pending == 0 {
				return fmt.Sprintf("%q will be marked done.", t)
			}
			return fmt.Sprintf("%q will be marked done. %d pending task(s) will be completed.", t, pending)
		},
		affirmative: "Complete",
		apply: func(svc gtd.ProjectService, ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
			return svc.CompleteProject(ctx, id, true, at)
		},
	},
	Drop: {
		title: "Drop project?",
		description: func(t string, pending int) string {
			if pending == 0 {
				return fmt.Sprintf("%q will be dropped.", t)
			}
			return fmt.Sprintf("%q will be dropped. %d pending task(s) will be dropped.", t, pending)
		},
		affirmative: "Drop",
		apply: func(svc gtd.ProjectService, ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
			return svc.DropProject(ctx, id, true, at)
		},
	},
}

type Model struct {
	project    gtd.Project
	svc        gtd.ProjectService
	transition Transition
	confirm    *bool
	at         **time.Time
	err        error
	form       *huh.Form
	applying   bool
}

func New(project gtd.Project, pending int, svc gtd.ProjectService, transition Transition) Model {
	m := Model{
		project:    project,
		svc:        svc,
		transition: transition,
		confirm:    new(true),
		at:         new(new(time.Now())),
	}

	s := specs[transition]
	when := date.NewField().
		Title("When").
		Value(m.at)
	confirm := huh.NewConfirm().
		Title(s.title).
		Description(s.description(project.Title, pending)).
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

func (m Model) Transition() Transition {
	return m.transition
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case projectTransitionedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.applying = false
			err := msg.err
			return m, func() tea.Msg { return fmt.Errorf("transition failed: %w", err) }
		}
		return m, screen.Dismiss()
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
		return m, tea.Batch(cmd, screen.Dismiss())
	case huh.StateCompleted:
		if !*m.confirm {
			return m, tea.Batch(cmd, screen.Dismiss())
		}
		m.applying = true
		return m, tea.Batch(cmd, m.applyCmd())
	}
	return m, cmd
}

func (m Model) applyCmd() tea.Cmd {
	id := m.project.ID
	svc := m.svc
	apply := specs[m.transition].apply
	at := time.Now()
	if m.at != nil && *m.at != nil {
		at = **m.at
	}
	return func() tea.Msg {
		_, err := apply(svc, context.Background(), id, at)
		if err != nil {
			slog.Error("transitioning project: " + err.Error())
		}
		return projectTransitionedMsg{err: err}
	}
}

func (m Model) View() string {
	return m.form.View()
}

func (m Model) CapturingInput() bool {
	return m.form.State == huh.StateNormal
}

func (m Model) KeyMap() help.KeyMap {
	return keyMap{m.form.KeyBinds()}
}

type projectTransitionedMsg struct {
	err error
}

type keyMap struct {
	binds []key.Binding
}

func (k keyMap) ShortHelp() []key.Binding { return k.binds }
func (k keyMap) FullHelp() [][]key.Binding {
	if len(k.binds) == 0 {
		return nil
	}
	return [][]key.Binding{k.binds}
}