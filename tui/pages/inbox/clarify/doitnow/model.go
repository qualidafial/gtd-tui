// Package doitnow renders the GTD do-it-now confirmation prompt: a small
// overlay shown after the wizard creates an open task that the user said they
// would do immediately. On Enter the task is marked done via the existing
// TaskService.CompleteTask path; on Esc the overlay dismisses and the task
// stays open (no work is lost — the user can complete it from the task list
// later). The overlay emits a ResultMsg alongside the dismiss so the wizard
// can decide whether to loop (project branch) or exit (single-task branch).
package doitnow

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
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	bannerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	descStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
)

// ResultMsg is emitted alongside the dismiss when the user closes the prompt.
// Completed=true means the task was marked done; Completed=false means the
// user left the task open (pressed Esc).
type ResultMsg struct {
	TaskID    int64
	Completed bool
}

type Model struct {
	task       gtd.Task
	svc        gtd.TaskService
	KeyMap     KeyMap
	completing bool
	err        error
}

func New(task gtd.Task, svc gtd.TaskService) Model {
	return Model{task: task, svc: svc, KeyMap: DefaultKeyMap()}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, m.KeyMap.Back) {
			m.err = nil
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case completedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.completing = false
			err := msg.err
			return m, cmds.Emit(fmt.Errorf("complete task: %w", err))
		}
		return screen.Dismiss(cmds.Emit(ResultMsg{TaskID: m.task.ID, Completed: true}))
	case tea.KeyPressMsg:
		if m.completing {
			return m, nil
		}
		switch {
		case key.Matches(msg, m.KeyMap.Confirm):
			m.completing = true
			return m, m.completeCmd()
		case key.Matches(msg, m.KeyMap.Back):
			return screen.Dismiss(cmds.Emit(ResultMsg{TaskID: m.task.ID, Completed: false}))
		}
	}
	return m, nil
}

func (m Model) completeCmd() tea.Cmd {
	svc := m.svc
	id := m.task.ID
	return func() tea.Msg {
		_, err := svc.CompleteTask(context.Background(), id, time.Now())
		if err != nil {
			slog.Error("complete task: " + err.Error())
		}
		return completedMsg{err: err}
	}
}

func (m Model) View() string {
	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left,
			bannerStyle.Render("Could not complete the task."),
			dimStyle.Render(m.err.Error()),
			"",
			dimStyle.Render("press esc to dismiss"),
		)
	}
	sections := []string{
		bannerStyle.Render("Do it now."),
		"",
		titleStyle.Render(m.task.Title),
	}
	if m.task.Description != "" {
		sections = append(sections, descStyle.Render(m.task.Description))
	}
	sections = append(sections,
		"",
		dimStyle.Render("press enter when finished  ·  esc to leave open"),
	)
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) Chords() []keymap.Group { return m.KeyMap.Chords() }

type completedMsg struct{ err error }
