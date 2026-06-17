package taskedit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/reltime"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/datefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/radiofield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/textfield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/theme"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

var (
	metaLabelStyle = theme.Label.Width(8)
	metaValueStyle = theme.Value
)

// ViewFactory builds the screen shown after a new task is created, so the
// editor can navigate to the new task's view. When nil, creating dismisses
// the editor without pushing a view (and updates always dismiss only).
type ViewFactory func(task gtd.Task) screen.Screen

type Model struct {
	task        gtd.Task
	svc         gtd.TaskService
	projectName string
	viewFactory ViewFactory
	err         error
	form        form.Model
	saving      bool
}

func New(task gtd.Task, svc gtd.TaskService, projectName string, viewFactory ViewFactory) Model {
	creating := task.ID == 0

	assignee := ""
	if task.Assignee != nil {
		assignee = *task.Assignee
	}

	title := inputfield.New("title", "Title",
		inputfield.WithValue(task.Title),
		inputfield.WithValidator(func(s string) error {
			if len(s) == 0 {
				return errors.New("title is required")
			}
			return nil
		}),
	)
	desc := textfield.New("description", "Description",
		textfield.WithValue(task.Description),
	)
	asg := inputfield.New("assignee", "Assignee",
		inputfield.WithValue(assignee),
	)

	dueOpts := []datefield.Option{}
	if task.Due != nil {
		dueOpts = append(dueOpts, datefield.WithValue(task.Due))
	}
	due := datefield.New("due", "Due", dueOpts...)

	deferOpts := []datefield.Option{}
	if task.DeferUntil != nil {
		deferOpts = append(deferOpts, datefield.WithValue(task.DeferUntil))
	}
	deferUntil := datefield.New("defer", "Defer Until", deferOpts...)

	// New tasks pick their initial status inline (Open or Done) on the
	// terminal field, so a finished task can be recorded straight to done.
	// Existing tasks keep a plain Save button; their status changes only
	// through the transition overlay.
	var terminal form.Field
	if creating {
		terminal = radiofield.New("status", "Status",
			[]radiofield.Option[gtd.TaskStatus]{
				{Display: "Open", Value: gtd.TaskStatusOpen},
				{Display: "Done", Value: gtd.TaskStatusDone},
			},
			radiofield.WithInitialValue(gtd.TaskStatusOpen),
		)
	} else {
		terminal = savefield.New("save", savefield.WithLabel("Save"))
	}

	return Model{
		task:        task,
		svc:         svc,
		projectName: projectName,
		viewFactory: viewFactory,
		form:        form.New(title, desc, asg, due, deferUntil, terminal),
	}
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
			m.err = nil
		}
		return m, nil
	}

	if m.saving {
		if sm, ok := msg.(taskSavedMsg); ok {
			return m.handleSaved(sm)
		}
		return m, nil
	}

	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	switch msg := msg.(type) {
	case form.SubmittedMsg:
		_ = msg
		m.saving = true
		return m, m.saveCmd()
	case taskSavedMsg:
		return m.handleSaved(msg)
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) handleSaved(msg taskSavedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("save failed: %w", err))
	}
	if msg.created && m.viewFactory != nil {
		return screen.Replace(m.viewFactory(msg.task))
	}
	return screen.Dismiss()
}

func (m Model) saveCmd() tea.Cmd {
	values := m.form.FieldValues()
	task := m.task
	task.Title, _ = values["title"].(string)
	task.Description, _ = values["description"].(string)
	if asg, _ := values["assignee"].(string); asg != "" {
		task.Assignee = new(asg)
	} else {
		task.Assignee = nil
	}
	task.Due, _ = values["due"].(*time.Time)
	task.DeferUntil, _ = values["defer"].(*time.Time)

	svc := m.svc
	creating := task.ID == 0
	if creating {
		// The create form's terminal field is the status radio; existing
		// tasks have no "status" key so their status is left untouched.
		if status, ok := values["status"].(gtd.TaskStatus); ok {
			task.Status = status
		}
	}
	return func() tea.Msg {
		var saved gtd.Task
		var err error
		ctx := context.Background()
		if creating {
			saved, err = svc.CreateTask(ctx, task)
		} else {
			saved, err = svc.UpdateTask(ctx, task)
		}
		if err != nil {
			slog.Error("saving task: " + err.Error())
		}
		return taskSavedMsg{task: saved, created: creating, err: err}
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
		)
		if m.projectName != "" {
			sections = append(sections, m.metaLine("Project", m.projectName))
		}
		sections = append(sections, "")
	}
	sections = append(sections, m.form.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) metaLine(label, value string) string {
	return metaLabelStyle.Render(label+":") + " " + metaValueStyle.Render(value)
}

func (m Model) statusValue() string {
	when := reltime.Format(m.task.StatusChangedAt, time.Now())
	return fmt.Sprintf("%s (%s)", titleStatus(m.task.Status), when)
}

func titleStatus(s gtd.TaskStatus) string {
	str := string(s)
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

func (m Model) CapturingInput() bool { return m.err == nil && !m.saving }

// Keys aggregates the form's resolved bindings and appends this screen's
// own esc/back binding as a trailing group. The overlay's esc is
// subtracted by Resolve since this screen claims esc.
func (m Model) Keys() []keymap.Group {
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}

type taskSavedMsg struct {
	task    gtd.Task
	created bool
	err     error
}
