package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
)

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestProjectTaskService_ListTasks(t *testing.T) {
	db := openTestDB(t)
	ctx := t.Context()

	taskSvc := service.NewTaskService(db)
	projSvc := service.NewProjectService(db)

	p1, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	p2, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P2", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)
	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "T2", Status: gtd.TaskStatusOpen, ProjectID: &p2.ID})
	require.NoError(t, err)
	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "T3", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	wrapped := service.NewProjectTaskService(taskSvc, p1.ID)
	tasks, err := wrapped.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)

	assert.Len(t, tasks, 1)
	assert.Equal(t, "T1", tasks[0].Title)
}

func TestProjectTaskService_CreateTask(t *testing.T) {
	db := openTestDB(t)
	ctx := t.Context()

	taskSvc := service.NewTaskService(db)
	projSvc := service.NewProjectService(db)

	p, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	wrapped := service.NewProjectTaskService(taskSvc, p.ID)
	created, err := wrapped.CreateTask(ctx, gtd.Task{Title: "New", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	assert.Equal(t, &p.ID, created.ProjectID)
}

func TestProjectTaskService_UpdateTask_Delegates(t *testing.T) {
	db := openTestDB(t)
	ctx := t.Context()

	taskSvc := service.NewTaskService(db)
	projSvc := service.NewProjectService(db)

	p, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "T1", Status: gtd.TaskStatusOpen, ProjectID: &p.ID})
	require.NoError(t, err)

	wrapped := service.NewProjectTaskService(taskSvc, p.ID)

	task.Title = "Updated"
	updated, err := wrapped.UpdateTask(ctx, task)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, &p.ID, updated.ProjectID)
}

func TestProjectTaskService_MoveTaskDown_StaysInProject(t *testing.T) {
	db := openTestDB(t)
	ctx := t.Context()

	taskSvc := service.NewTaskService(db)
	projSvc := service.NewProjectService(db)

	p1, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	p2, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P2", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	a, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "a", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)
	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "x", Status: gtd.TaskStatusOpen, ProjectID: &p2.ID})
	require.NoError(t, err)
	b, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "b", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)

	// Caller passes an empty filter; the wrapper must inject ProjectID=p1
	// so the move stays scoped to P1's tasks.
	wrapped := service.NewProjectTaskService(taskSvc, p1.ID)
	require.NoError(t, wrapped.MoveTaskDown(ctx, a.ID, gtd.TaskFilter{}))

	tasks, err := wrapped.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	gotIDs := make([]int64, len(tasks))
	for i, task := range tasks {
		gotIDs[i] = task.ID
	}
	assert.Equal(t, []int64{b.ID, a.ID}, gotIDs)
}

func TestProjectTaskService_MoveTaskUp_OverridesForeignProjectID(t *testing.T) {
	db := openTestDB(t)
	ctx := t.Context()

	taskSvc := service.NewTaskService(db)
	projSvc := service.NewProjectService(db)

	p1, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	p2, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P2", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "a", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)
	b, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "b", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)

	// Caller passes a foreign ProjectID; wrapper must overwrite it with p1.
	wrapped := service.NewProjectTaskService(taskSvc, p1.ID)
	require.NoError(t, wrapped.MoveTaskUp(ctx, b.ID, gtd.TaskFilter{ProjectID: &p2.ID}))

	tasks, err := wrapped.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	assert.Equal(t, "b", tasks[0].Title)
	assert.Equal(t, "a", tasks[1].Title)
}

func TestProjectTaskService_MoveTaskFirstLast_StaysInProject(t *testing.T) {
	db := openTestDB(t)
	ctx := t.Context()

	taskSvc := service.NewTaskService(db)
	projSvc := service.NewProjectService(db)

	p1, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	p2, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P2", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	// Interleave P1/P2 tasks so a project-wide move would reshuffle foreign
	// tasks if the wrapper failed to scope the filter to P1.
	a, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "a", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)
	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "x", Status: gtd.TaskStatusOpen, ProjectID: &p2.ID})
	require.NoError(t, err)
	b, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "b", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)
	cTask, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "c", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	require.NoError(t, err)

	// Caller passes a foreign ProjectID; the wrapper must overwrite it with p1.
	wrapped := service.NewProjectTaskService(taskSvc, p1.ID)
	require.NoError(t, wrapped.MoveTaskLast(ctx, a.ID, gtd.TaskFilter{ProjectID: &p2.ID}))

	gotIDs := func() []int64 {
		tasks, err := wrapped.ListTasks(ctx, gtd.TaskFilter{})
		require.NoError(t, err)
		ids := make([]int64, len(tasks))
		for i, task := range tasks {
			ids[i] = task.ID
		}
		return ids
	}
	// Within P1: [b, c, a].
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, gotIDs())

	require.NoError(t, wrapped.MoveTaskFirst(ctx, cTask.ID, gtd.TaskFilter{}))
	// Within P1: [c, b, a].
	assert.Equal(t, []int64{cTask.ID, b.ID, a.ID}, gotIDs())

	// P2's lone task is untouched throughout.
	p2Tasks, err := taskSvc.ListTasks(ctx, gtd.TaskFilter{ProjectID: &p2.ID})
	require.NoError(t, err)
	require.Len(t, p2Tasks, 1)
	assert.Equal(t, "x", p2Tasks[0].Title)
}

func TestProjectTaskService_ListTasks_WithCallerFilter(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	taskSvc := service.NewTaskService(db)
	projSvc := service.NewProjectService(db)

	p, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "Pending", Status: gtd.TaskStatusOpen, ProjectID: &p.ID})
	require.NoError(t, err)
	done := gtd.TaskStatusDone
	_, err = taskSvc.CreateTask(ctx, gtd.Task{Title: "Done", Status: gtd.TaskStatusDone, ProjectID: &p.ID})
	require.NoError(t, err)

	wrapped := service.NewProjectTaskService(taskSvc, p.ID)
	tasks, err := wrapped.ListTasks(ctx, gtd.TaskFilter{Status: &done})
	require.NoError(t, err)

	assert.Len(t, tasks, 1)
	assert.Equal(t, "Done", tasks[0].Title)
}