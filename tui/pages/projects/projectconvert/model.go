// Package projectconvert is the confirm overlay for collapsing an empty open
// project into a standalone task. It is confirm-only: on confirm it broadcasts
// ConfirmedMsg to the calling screen and dismisses, applying no change itself.
// The caller (project view or project list) performs ConvertProjectToTask and
// handles navigation, mirroring the selection-only taskpicker pattern.
package projectconvert

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var (
	keyBack    = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	titleStyle = lipgloss.NewStyle().Bold(true)
	descStyle  = lipgloss.NewStyle().Faint(true)
)

// ConfirmedMsg signals the user confirmed the conversion. The calling screen
// applies ConvertProjectToTask for ProjectID once it is active again.
type ConfirmedMsg struct {
	ProjectID int64
}

type Model struct {
	project gtd.Project
	form    form.Model
}

func New(project gtd.Project) Model {
	return Model{
		project: project,
		form:    form.New(savefield.New("save", savefield.WithLabel("Convert"))),
	}
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	if _, ok := msg.(form.SubmittedMsg); ok {
		return screen.Dismiss(cmds.Emit(ConfirmedMsg{ProjectID: m.project.ID}))
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	header := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Convert to task?"),
		descStyle.Render(fmt.Sprintf("%q will become a standalone task. The project will be removed.", m.project.Title)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.form.View())
}

func (m Model) CapturingInput() bool { return true }

// Keys aggregates the form's resolved bindings and appends this screen's own
// esc binding as a trailing group; Resolve subtracts the overlay's duplicate
// esc.
func (m Model) Keys() []keymap.Group {
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}
