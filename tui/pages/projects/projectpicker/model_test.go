package projectpicker

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
)

type env struct {
	taskSvc    gtd.TaskService
	projectSvc gtd.ProjectService
}

func setup(t *testing.T) env {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return env{
		taskSvc:    service.NewTaskService(db),
		projectSvc: service.NewProjectService(db),
	}
}

func TestPicker_Assign(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	p, err := e.projectSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)
	assert.Nil(t, task.ProjectID)

	var m screen.Screen = New(task, e.taskSvc, e.projectSvc)
	m = screentest.Init(t, m)

	// Select the project (index 1, since 0 is "(none)")
	m = screentest.Send(t, m, tea.KeyPressMsg{Code: tea.KeyDown})

	var dismissed bool
	for s, msg := range screentest.PumpSend(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) {
		m = s
		if _, ok := msg.(screen.DismissMsg); ok {
			dismissed = true
			break
		}
	}
	require.True(t, dismissed, "enter should dismiss after save")

	updated, err := e.taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.ProjectID)
	assert.Equal(t, p.ID, *updated.ProjectID)
}

func TestPicker_Unlink(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	p, err := e.projectSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Status: gtd.TaskStatusOpen, ProjectID: &p.ID})
	require.NoError(t, err)

	var m screen.Screen = New(task, e.taskSvc, e.projectSvc)
	m = screentest.Init(t, m)

	// Move to "(none)" and confirm
	for s := range screentest.PumpSend(t, m, tea.KeyPressMsg{Code: tea.KeyHome}) {
		m = s
	}

	var dismissed bool
	for s, msg := range screentest.PumpSend(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) {
		m = s
		if _, ok := msg.(screen.DismissMsg); ok {
			dismissed = true
			break
		}
	}
	require.True(t, dismissed, "enter should dismiss after save")

	updated, err := e.taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.ProjectID)
}

func TestPicker_NoChange_SkipsUpdate(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	p, err := e.projectSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Status: gtd.TaskStatusOpen, ProjectID: &p.ID})
	require.NoError(t, err)

	var m screen.Screen = New(task, e.taskSvc, e.projectSvc)
	m = screentest.Init(t, m)

	// Submit without changing selection
	var dismissed bool
	for s, msg := range screentest.PumpSend(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) {
		m = s

		switch msg.(type) {
		case savedMsg:
			t.Fatalf("picker should not have attempted to save")
		case screen.DismissMsg:
			dismissed = true
			break
		}
	}
	require.True(t, dismissed, "enter with no change should dismiss without saving")

	got, err := e.taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, &p.ID, got.ProjectID)
}

func TestPicker_Cancel(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	var m screen.Screen = New(task, e.taskSvc, e.projectSvc)
	m = screentest.Init(t, m)

	var dismissed bool
	for s, msg := range screentest.PumpSend(t, m, tea.KeyPressMsg{Code: tea.KeyEscape}) {
		m = s
		if _, ok := msg.(screen.DismissMsg); ok {
			dismissed = true
			break
		}
	}
	require.True(t, dismissed, "esc should dismiss")
}
