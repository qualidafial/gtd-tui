package taskedit

import (
	"context"
	"errors"
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
	"github.com/qualidafial/gtd-tui/tui/pages/tasks"
)

var (
	taskStatusOptions = []huh.Option[gtd.TaskStatus]{
		huh.NewOption("Inbox", gtd.TaskStatusInbox),
		huh.NewOption("Active", gtd.TaskStatusActive),
		huh.NewOption("Waiting", gtd.TaskStatusWaiting),
		huh.NewOption("Deferred", gtd.TaskStatusDeferred),
		huh.NewOption("Done", gtd.TaskStatusDone),
		huh.NewOption("Dropped", gtd.TaskStatusDropped),
	}
)

type Model struct {
	task   *gtd.Task
	svc    gtd.TaskService
	err    error
	form   *huh.Form
	saving bool
}

func New(task gtd.Task, svc gtd.TaskService) Model {
	m := Model{
		task: &task,
		svc:  svc,
	}

	var fields []huh.Field
	if task.ID != 0 {
		fields = append(fields,
			huh.NewNote().
				Title("Task ID").
				Description(fmt.Sprint(task.ID)),
			huh.NewNote().
				Title("Created").
				Description(task.CreatedAt.Local().Format(time.DateTime)),
			huh.NewNote().
				Title("Updated").
				Description(task.UpdatedAt.Local().Format(time.DateTime)),
		)
	}
	fields = append(fields,
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
		huh.NewSelect[gtd.TaskStatus]().
			Title("Status").
			Options(taskStatusOptions...).
			Value(&task.Status),
		date.NewField().
			Title("Due").
			Value(&task.Due),
		date.NewField().
			Title("Defer Until").
			Value(&task.DeferUntil),
	)
	group := huh.NewGroup(fields...)

	// Extend the form's Quit binding so esc aborts in addition to ctrl+c.
	// ctrl+c is intercepted at app level for whole-program quit; here esc
	// aborts the form, which we translate into HideOverlay below.
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

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
		return m, tea.Batch(screen.HideOverlay(), tasks.TasksChanged())
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
		return m, tea.Batch(cmd, screen.HideOverlay())
	case huh.StateCompleted:
		m.saving = true
		return m, tea.Batch(cmd, m.saveCmd())
	}
	return m, cmd
}

func (m Model) saveCmd() tea.Cmd {
	task := *m.task
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
	return m.form.View()
}

func (m Model) KeyMap() help.KeyMap {
	return KeyMap{m.form.KeyBinds()}
}

type taskSavedMsg struct {
	task gtd.Task
	err  error
}
