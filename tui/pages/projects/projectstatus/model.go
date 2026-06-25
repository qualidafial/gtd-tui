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
	"github.com/qualidafial/gtd-tui/tui/components/form/selectfield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/theme"
)

// Transition identifies a project status change initiated from the
// project list.
type Transition int

const (
	Complete Transition = iota
	Drop
	Park
	Reopen
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
	Park: {
		title: "Park project?",
		description: func(t string, _ int) string {
			return fmt.Sprintf("%q will be moved to Someday. Task statuses are unchanged.", t)
		},
		affirmative: "Park",
		apply: func(svc gtd.ProjectService, ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
			return svc.ParkProject(ctx, id, at)
		},
	},
	Reopen: {
		title: "Reopen project?",
		description: func(t string, _ int) string {
			return fmt.Sprintf("%q will be moved back to open. Task statuses are unchanged.", t)
		},
		affirmative: "Reopen",
		apply: func(svc gtd.ProjectService, ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
			return svc.ReopenProject(ctx, id, at)
		},
	},
}

const (
	fieldStatus = "status"
	fieldWhen   = "when"
	fieldSave   = "save"
)

var (
	keyBack    = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	titleStyle = theme.Title
	descStyle  = theme.Subtitle
)

// Model is the project status overlay. It runs in one of two modes: a fixed
// single-transition confirmation (New, used by the delete fast-drop) or a
// status picker (NewPicker) that lets the user choose any reachable status.
// Both share the editable When timestamp and apply step.
type Model struct {
	project    gtd.Project
	pending    int
	svc        gtd.ProjectService
	transition Transition // applied in fixed mode (!pick)
	pick       bool
	form       form.Model
	applying   bool
}

// New builds a fixed-transition confirmation overlay: an editable When date
// and a Save button. Used by the delete fast-drop shortcut.
func New(project gtd.Project, pending int, svc gtd.ProjectService, transition Transition) Model {
	s := specs[transition]
	now := time.Now()

	when := datefield.New(fieldWhen, "When", datefield.WithValue(&now))
	save := savefield.New(fieldSave, savefield.WithLabel(s.affirmative))

	return Model{
		project:    project,
		pending:    pending,
		svc:        svc,
		transition: transition,
		form:       form.New(when, save),
	}
}

// NewPicker builds the status picker: a status selectfield (current status
// preselected) followed by an editable When date that appears only once a
// different status is chosen, then a Save button. ctrl+s saves from anywhere;
// confirming on the unchanged status is a no-op that dismisses. pending is the
// count of not-yet-complete tasks, surfaced as cascade info for Complete/Drop.
func NewPicker(project gtd.Project, pending int, svc gtd.ProjectService) Model {
	now := time.Now()
	opts := optionsFor(project.Status)
	// The list reserves one row, so size it to len+1 to show every status.
	status := selectfield.New(fieldStatus, "Status", opts,
		selectfield.WithInitialValue(project.Status),
		selectfield.WithHeight[gtd.ProjectStatus](len(opts)+1))
	when := datefield.New(fieldWhen, "When", datefield.WithValue(&now),
		datefield.WithVisible(func(v form.Values) bool {
			return v.Get(fieldStatus) != any(project.Status)
		}))
	save := savefield.New(fieldSave, savefield.WithLabel("Save"))

	return Model{
		project: project,
		pending: pending,
		svc:     svc,
		pick:    true,
		form:    form.New(status, when, save),
	}
}

// Transition reports the transition this overlay will apply: the fixed one in
// confirmation mode, or the one implied by the current picker selection.
func (m Model) Transition() Transition {
	if !m.pick {
		return m.transition
	}
	t, _ := transitionFor(m.project.Status, m.selectedStatus())
	return t
}

// Current returns the subject project's current status (the preselection).
func (m Model) Current() gtd.ProjectStatus { return m.project.Status }

// Picking reports whether this is the multi-status picker (vs a fixed
// single-transition confirmation).
func (m Model) Picking() bool { return m.pick }

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
		// In picker mode, confirming the unchanged status applies nothing.
		if m.pick && m.selectedStatus() == m.project.Status {
			return screen.Dismiss()
		}
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

// applyCmd issues the service transition with the confirmed When instant. The
// transition is the fixed one in confirmation mode, or the one mapped from the
// chosen status in picker mode.
func (m Model) applyCmd() tea.Cmd {
	transition := m.transition
	if m.pick {
		t, ok := transitionFor(m.project.Status, m.selectedStatus())
		if !ok {
			return nil
		}
		transition = t
	}
	id := m.project.ID
	svc := m.svc
	apply := specs[transition].apply
	at := m.whenValue()
	return func() tea.Msg {
		_, err := apply(svc, context.Background(), id, at)
		if err != nil {
			slog.Error("transitioning project: " + err.Error())
		}
		return projectTransitionedMsg{err: err}
	}
}

// selectedStatus returns the status chosen in the picker, or the project's
// current status when there is no status field (confirmation mode).
func (m Model) selectedStatus() gtd.ProjectStatus {
	if s, ok := m.form.FieldValues()[fieldStatus].(gtd.ProjectStatus); ok {
		return s
	}
	return m.project.Status
}

// whenValue reads the When field, falling back to now if it is empty or hidden.
func (m Model) whenValue() time.Time {
	if t, _ := m.form.FieldValues()[fieldWhen].(*time.Time); t != nil {
		return *t
	}
	return time.Now()
}

func (m Model) View() string {
	lines := []string{titleStyle.Render(m.title())}
	// The picker keeps a stable layout: the chosen status and the When field
	// already convey the action, so no per-selection description is shown (a
	// description that appears/disappears would shift the form and reorient the
	// eye). The fixed-transition confirmation keeps its static description.
	if !m.pick {
		lines = append(lines, descStyle.Render(specs[m.transition].description(m.project.Title, m.pending)))
	}
	header := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.form.View())
}

// title is the overlay heading for the current mode.
func (m Model) title() string {
	if m.pick {
		return "Project status"
	}
	return specs[m.transition].title
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
