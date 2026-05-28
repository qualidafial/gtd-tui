package taskedit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/reltime"
	"github.com/qualidafial/gtd-tui/tui/components/date"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

var (
	metaLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(8)
	metaValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)

type Model struct {
	task     *gtd.Task
	svc      gtd.TaskService
	assignee string
	err      error
	form     *huh.Form
	saving   bool
}

func New(task gtd.Task, svc gtd.TaskService) Model {
	if task.ID == 0 {
		task.Status = gtd.TaskStatusOpen
	}

	assignee := ""
	if task.Assignee != nil {
		assignee = *task.Assignee
	}

	m := Model{
		task:     &task,
		svc:      svc,
		assignee: assignee,
	}

	fields := []huh.Field{
		huh.NewInput().
			Title("Title").
			Value(&task.Title).
			Validate(func(s string) error {
				if len(s) == 0 {
					return errors.New("title is required")
				}
				return nil
			}),
		huh.NewText().
			Title("Description").
			Value(&task.Description),
		huh.NewInput().
			Title("Assignee").
			Value(&m.assignee),
		date.NewField().
			Title("Due").
			Value(&task.Due),
		date.NewField().
			Title("Defer Until").
			Value(&task.DeferUntil),
	}
	group := huh.NewGroup(fields...)

	// Extend the form's Quit binding so esc aborts in addition to ctrl+c.
	// ctrl+c is intercepted at app level for whole-program quit; here esc
	// aborts the form, which we translate into HideOverlay below.
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = keyBack

	m.form = huh.NewForm(group).
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
	case taskSavedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.saving = false
			return m, nil
		}
		return m, screen.Dismiss()
	case tea.KeyPressMsg:
		// After a save error the form is stuck in StateCompleted, so
		// any key that fell through to the form would re-trigger the
		// save loop. Esc clears the error and rewinds the form to
		// StateNormal so the user can edit and retry; other keys are
		// swallowed until the error is cleared.
		if m.err != nil {
			if key.Matches(msg, keyBack) {
				m.err = nil
				m.form.State = huh.StateNormal
			}
			return m, nil
		}
	}

	if m.saving {
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
		m.saving = true
		return m, tea.Batch(cmd, m.saveCmd())
	}
	return m, cmd
}

func (m Model) saveCmd() tea.Cmd {
	task := *m.task
	if m.assignee != "" {
		task.Assignee = new(m.assignee)
	} else {
		task.Assignee = nil
	}
	svc := m.svc
	return func() tea.Msg {
		var saved gtd.Task
		var err error
		ctx := context.Background()
		if task.ID == 0 {
			saved, err = svc.CreateTask(ctx, task)
		} else {
			saved, err = svc.UpdateTask(ctx, task)
		}
		if err != nil {
			slog.Error("saving task: " + err.Error())
		}
		return taskSavedMsg{
			task: saved,
			err:  err,
		}
	}
}

func (m Model) View() string {
	var sections []string
	if m.task.ID != 0 {
		sections = append(sections,
			m.metaLine("Task ID", fmt.Sprint(m.task.ID)),
			m.metaLine("Created", m.task.CreatedAt.Local().Format(time.DateTime)),
			m.metaLine("Updated", m.task.UpdatedAt.Local().Format(time.DateTime)),
			m.metaLine("Status", m.statusValue()),
			"",
		)
	}
	sections = append(sections, m.form.View())
	if m.err != nil {
		sections = append(sections, "", errorStyle.Render("save failed: "+m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) metaLine(label, value string) string {
	return metaLabelStyle.Render(label+":") + " " + metaValueStyle.Render(value)
}

// statusValue renders the status name (first letter capitalized) followed by a
// relative WHEN of the last status change: "Pending (3d)", "Done (today)".
func (m Model) statusValue() string {
	when := reltime.Format(m.task.StatusChangedAt, time.Now())
	return fmt.Sprintf("%s (%s)", titleStatus(m.task.Status), when)
}

// titleStatus maps a lowercase TaskStatus to a display form with its first
// letter capitalized (e.g. "open" -> "Open").
func titleStatus(s gtd.TaskStatus) string {
	str := string(s)
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

func (m Model) CapturingInput() bool {
	return m.form.State == huh.StateNormal
}

func (m Model) KeyMap() help.KeyMap {
	return KeyMap{m.form.KeyBinds()}
}

type taskSavedMsg struct {
	task gtd.Task
	err  error
}
