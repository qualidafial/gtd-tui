package projectstatus

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
)

// Transition identifies a project status change initiated from the
// project list.
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

var (
	keyBack    = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	titleStyle = lipgloss.NewStyle().Bold(true)
	descStyle  = lipgloss.NewStyle().Faint(true)
)

type Model struct {
	project    gtd.Project
	pending    int
	svc        gtd.ProjectService
	transition Transition
	form       form.Model
	applying   bool
}

func New(project gtd.Project, pending int, svc gtd.ProjectService, transition Transition) Model {
	s := specs[transition]
	now := time.Now()

	when := datefield.New("when", "When", datefield.WithValue(&now))
	save := savefield.New("save", savefield.WithLabel(s.affirmative))

	return Model{
		project:    project,
		pending:    pending,
		svc:        svc,
		transition: transition,
		form:       form.New(when, save),
	}
}

func (m Model) Transition() Transition { return m.transition }

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if m.applying {
		if tm, ok := msg.(projectTransitionedMsg); ok {
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
	case projectTransitionedMsg:
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

func (m Model) applyCmd() tea.Cmd {
	id := m.project.ID
	svc := m.svc
	apply := specs[m.transition].apply
	at := time.Now()
	if t, _ := m.form.FieldValues()["when"].(*time.Time); t != nil {
		at = *t
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
	s := specs[m.transition]
	header := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(s.title),
		descStyle.Render(s.description(m.project.Title, m.pending)),
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

type projectTransitionedMsg struct {
	err error
}
