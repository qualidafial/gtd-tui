package projectpicker

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
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

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	require.NoError(t, err)
	assert.Nil(t, task.ProjectID)

	m := New(task, e.taskSvc, e.projectSvc)

	// Simulate load
	m = applyMsg(t, m, m.loadCmd()())

	// Select the project (index 1, since 0 is "(none)")
	m.selected = new(p.ID)
	m.form.State = huh.StateCompleted

	_, cmd := m.Update(nil)
	msg := cmd()
	m = applyMsg(t, m, msg)

	// Verify dismiss
	_, cmd2 := m.Update(msg)
	require.NotNil(t, cmd2)
	dismissMsg := cmd2()
	_, ok := dismissMsg.(screen.DismissMsg)
	assert.True(t, ok)

	// Verify task was updated
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

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending, ProjectID: &p.ID})
	require.NoError(t, err)

	m := New(task, e.taskSvc, e.projectSvc)
	m = applyMsg(t, m, m.loadCmd()())

	// Select "(none)"
	m.selected = nil
	m.form.State = huh.StateCompleted

	_, cmd := m.Update(nil)
	msg := cmd()
	m = applyMsg(t, m, msg)

	updated, err := e.taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.ProjectID)
}

func TestPicker_NoChange_SkipsUpdate(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	p, err := e.projectSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending, ProjectID: &p.ID})
	require.NoError(t, err)

	m := New(task, e.taskSvc, e.projectSvc)
	m = applyMsg(t, m, m.loadCmd()())

	// Keep the same project selected (original == selected)
	m.form.State = huh.StateCompleted

	assert.False(t, m.saving, "should not have started saving")

	// Verify task is unchanged
	got, err := e.taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, &p.ID, got.ProjectID)
}

func TestPicker_Cancel(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	task, err := e.taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	require.NoError(t, err)

	m := New(task, e.taskSvc, e.projectSvc)
	m = applyMsg(t, m, m.loadCmd()())

	m.form.State = huh.StateAborted
	_, cmd := m.Update(nil)
	require.NotNil(t, cmd)
}

func applyMsg(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	updated, _ := m.Update(msg)
	return updated.(Model)
}
