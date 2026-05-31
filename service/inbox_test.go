package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
)

func TestInboxService_Discard(t *testing.T) {
	t.Run("marks item discarded", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "junk"})
		require.NoError(t, err)

		got, err := svc.Discard(ctx, item.ID)
		require.NoError(t, err)
		assert.True(t, got.Discarded)
		assert.True(t, got.UpdatedAt.After(item.UpdatedAt))

		// And the inbox no longer lists it.
		live, err := svc.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, live)
	})

	t.Run("rejects missing item", func(t *testing.T) {
		db := openTestDB(t)
		svc := service.NewInboxService(db)
		_, err := svc.Discard(t.Context(), 999)
		require.Error(t, err)
	})

	t.Run("rejects already-clarified item", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "x"})
		require.NoError(t, err)
		_, _, err = svc.ClarifyAsTask(ctx, item.ID, gtd.Task{})
		require.NoError(t, err)

		_, err = svc.Discard(ctx, item.ID)
		require.Error(t, err)
	})

	t.Run("rejects already-discarded item", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "x"})
		require.NoError(t, err)
		_, err = svc.Discard(ctx, item.ID)
		require.NoError(t, err)

		_, err = svc.Discard(ctx, item.ID)
		require.Error(t, err)
	})
}

func TestInboxService_Incubate(t *testing.T) {
	t.Run("creates someday project and stamps item", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)
		projSvc := service.NewProjectService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Learn pottery"})
		require.NoError(t, err)

		project, updated, err := svc.Incubate(ctx, item.ID, gtd.Project{})
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusSomeday, project.Status)
		assert.Equal(t, "Learn pottery", project.Title)
		require.NotNil(t, updated.ClarifiedIntoProjectID)
		assert.Equal(t, project.ID, *updated.ClarifiedIntoProjectID)

		// ReopenProject brings it back to open without further item changes.
		reopened, err := projSvc.ReopenProject(ctx, project.ID, project.StatusChangedAt)
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusOpen, reopened.Status)
	})

	t.Run("uses explicit title when provided", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "default"})
		require.NoError(t, err)
		project, _, err := svc.Incubate(ctx, item.ID, gtd.Project{Title: "explicit"})
		require.NoError(t, err)
		assert.Equal(t, "explicit", project.Title)
	})

	t.Run("inherits item description when blank", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Plan trip", Description: "Spring 2027"})
		require.NoError(t, err)

		project, _, err := svc.Incubate(ctx, item.ID, gtd.Project{})
		require.NoError(t, err)
		assert.Equal(t, "Spring 2027", project.Description)
	})

	t.Run("rejects already-clarified item", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "x"})
		require.NoError(t, err)
		_, _, _, err = svc.ClarifyAsProject(ctx, item.ID, gtd.Project{}, gtd.Task{Title: "step 1"})
		require.NoError(t, err)

		_, _, err = svc.Incubate(ctx, item.ID, gtd.Project{})
		require.Error(t, err)
	})
}

func TestInboxService_ClarifyAsTask(t *testing.T) {
	t.Run("creates next_action task by default", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Pay bill"})
		require.NoError(t, err)

		task, updated, err := svc.ClarifyAsTask(ctx, item.ID, gtd.Task{})
		require.NoError(t, err)
		assert.Equal(t, "Pay bill", task.Title)
		assert.Equal(t, gtd.TaskStatusOpen, task.Status)
		require.NotNil(t, updated.ClarifiedIntoTaskID)
		assert.Equal(t, task.ID, *updated.ClarifiedIntoTaskID)
	})

	t.Run("inherits item description when blank", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Send invoice", Description: "Attach receipt"})
		require.NoError(t, err)
		task, _, err := svc.ClarifyAsTask(ctx, item.ID, gtd.Task{})
		require.NoError(t, err)
		assert.Equal(t, "Attach receipt", task.Description)
	})

	t.Run("creates delegated task with assignee", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Review PR"})
		require.NoError(t, err)

		alice := "alice"
		task, _, err := svc.ClarifyAsTask(ctx, item.ID, gtd.Task{Assignee: &alice})
		require.NoError(t, err)
		require.NotNil(t, task.Assignee)
		assert.Equal(t, "alice", *task.Assignee)
	})

	t.Run("attaches task to project", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)
		projSvc := service.NewProjectService(db)

		proj, err := projSvc.CreateProject(ctx, gtd.Project{Title: "P", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		item, err := svc.Create(ctx, gtd.Item{Title: "step 1"})
		require.NoError(t, err)

		task, _, err := svc.ClarifyAsTask(ctx, item.ID, gtd.Task{ProjectID: &proj.ID})
		require.NoError(t, err)
		require.NotNil(t, task.ProjectID)
		assert.Equal(t, proj.ID, *task.ProjectID)
	})

	t.Run("rejects invalid project", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "x"})
		require.NoError(t, err)

		bogus := int64(999)
		_, _, err = svc.ClarifyAsTask(ctx, item.ID, gtd.Task{ProjectID: &bogus})
		require.Error(t, err)

		// Item must still be unclarified (rollback).
		got, err := svc.Get(ctx, item.ID)
		require.NoError(t, err)
		assert.Nil(t, got.ClarifiedIntoTaskID)
	})

	t.Run("rejects non-open status", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Already done"})
		require.NoError(t, err)

		// Do-it-now is wizard-orchestrated: clarify creates an open task, then
		// the wizard calls CompleteTask after the user confirms. The service
		// rejects any non-open status to keep the contract honest.
		_, _, err = svc.ClarifyAsTask(ctx, item.ID, gtd.Task{Status: gtd.TaskStatusDone})
		require.Error(t, err)
	})

	t.Run("rejects already-clarified item", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "x"})
		require.NoError(t, err)
		_, _, err = svc.ClarifyAsTask(ctx, item.ID, gtd.Task{})
		require.NoError(t, err)

		_, _, err = svc.ClarifyAsTask(ctx, item.ID, gtd.Task{})
		require.Error(t, err)
	})
}

func TestInboxService_ClarifyAsProject(t *testing.T) {
	t.Run("creates open project with one open first task (vs Incubate's someday)", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		open, err := svc.Create(ctx, gtd.Item{Title: "Now"})
		require.NoError(t, err)
		later, err := svc.Create(ctx, gtd.Item{Title: "Maybe"})
		require.NoError(t, err)

		nowProject, nowTask, _, err := svc.ClarifyAsProject(ctx, open.ID, gtd.Project{}, gtd.Task{Title: "first step"})
		require.NoError(t, err)
		laterProject, _, err := svc.Incubate(ctx, later.ID, gtd.Project{})
		require.NoError(t, err)

		assert.Equal(t, gtd.ProjectStatusOpen, nowProject.Status)
		assert.Equal(t, gtd.ProjectStatusSomeday, laterProject.Status)
		assert.Equal(t, gtd.TaskStatusOpen, nowTask.Status)
		require.NotNil(t, nowTask.ProjectID)
		assert.Equal(t, nowProject.ID, *nowTask.ProjectID)
	})

	t.Run("copies item title by default", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Plan move"})
		require.NoError(t, err)
		project, _, _, err := svc.ClarifyAsProject(ctx, item.ID, gtd.Project{}, gtd.Task{Title: "Book movers"})
		require.NoError(t, err)
		assert.Equal(t, "Plan move", project.Title)
	})

	t.Run("rejects non-open first task status", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "x"})
		require.NoError(t, err)

		// First task is the checkpoint; it must be open. Do-it-now is handled
		// by the wizard via CompleteTask after the user confirms.
		_, _, _, err = svc.ClarifyAsProject(ctx, item.ID, gtd.Project{}, gtd.Task{Title: "t", Status: gtd.TaskStatusDone})
		require.Error(t, err)

		// Item must still be unclarified.
		got, err := svc.Get(ctx, item.ID)
		require.NoError(t, err)
		assert.Nil(t, got.ClarifiedIntoProjectID)
	})

	t.Run("rejects caller-supplied first task ProjectID", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)
		projSvc := service.NewProjectService(db)

		other, err := projSvc.CreateProject(ctx, gtd.Project{Title: "Other", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		item, err := svc.Create(ctx, gtd.Item{Title: "x"})
		require.NoError(t, err)

		_, _, _, err = svc.ClarifyAsProject(ctx, item.ID, gtd.Project{}, gtd.Task{Title: "t", ProjectID: &other.ID})
		require.Error(t, err)
	})

	t.Run("rejects missing item", func(t *testing.T) {
		db := openTestDB(t)
		svc := service.NewInboxService(db)
		_, _, _, err := svc.ClarifyAsProject(t.Context(), 999, gtd.Project{}, gtd.Task{Title: "x"})
		require.Error(t, err)
	})

	t.Run("wizard checkpoint pattern: add second task then complete first via TaskService", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		svc := service.NewInboxService(db)
		taskSvc := service.NewTaskService(db)

		item, err := svc.Create(ctx, gtd.Item{Title: "Refactor auth"})
		require.NoError(t, err)

		project, firstTask, _, err := svc.ClarifyAsProject(ctx, item.ID, gtd.Project{Outcome: "auth is modular"}, gtd.Task{Title: "Sketch design"})
		require.NoError(t, err)

		// Wizard completes the first task after the user confirms do-it-now.
		completed, err := taskSvc.CompleteTask(ctx, firstTask.ID, firstTask.StatusChangedAt)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusDone, completed.Status)

		// Wizard then adds the next task via TaskService, pre-attached.
		next, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "Draft interface", Status: gtd.TaskStatusOpen, ProjectID: &project.ID})
		require.NoError(t, err)
		require.NotNil(t, next.ProjectID)
		assert.Equal(t, project.ID, *next.ProjectID)
		assert.Equal(t, gtd.TaskStatusOpen, next.Status)
	})
}

func TestInboxService_CaptureClarifyFlow(t *testing.T) {
	db := openTestDB(t)
	ctx := t.Context()
	svc := service.NewInboxService(db)

	// Capture three items.
	a, err := svc.Create(ctx, gtd.Item{Title: "A"})
	require.NoError(t, err)
	b, err := svc.Create(ctx, gtd.Item{Title: "B"})
	require.NoError(t, err)
	c, err := svc.Create(ctx, gtd.Item{Title: "C"})
	require.NoError(t, err)

	// All visible in FIFO.
	items, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, a.ID, items[0].ID)

	// Clarify each via a different path.
	_, _, err = svc.ClarifyAsTask(ctx, a.ID, gtd.Task{})
	require.NoError(t, err)
	_, _, _, err = svc.ClarifyAsProject(ctx, b.ID, gtd.Project{}, gtd.Task{Title: "first step"})
	require.NoError(t, err)
	_, err = svc.Discard(ctx, c.ID)
	require.NoError(t, err)

	// Inbox now empty.
	items, err = svc.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}
