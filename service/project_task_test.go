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