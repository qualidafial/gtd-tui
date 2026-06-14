package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/sqlite"
)

// --- 6. Project CRUD ---

func TestDB_CreateProject(t *testing.T) {
	t.Run("minimal project", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		got, err := db.CreateProject(c, gtd.Project{
			Title:  "Launch website",
			Status: gtd.ProjectStatusOpen,
		})
		require.NoError(t, err)

		assert.NotZero(t, got.ID)
		assert.False(t, got.CreatedAt.IsZero())
		assert.False(t, got.UpdatedAt.IsZero())
		assert.Equal(t, "Launch website", got.Title)
		assert.Equal(t, gtd.ProjectStatusOpen, got.Status)

		fetched, err := db.GetProject(c, got.ID)
		require.NoError(t, err)
		assert.Equal(t, got.Title, fetched.Title)
		assert.Equal(t, got.Status, fetched.Status)
	})

	t.Run("full project with due date", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		got, err := db.CreateProject(c, gtd.Project{
			Title:       "Launch website",
			Outcome:     "Website is live and accepting users",
			Description: "Covers design, dev, and deployment",
			Status:      gtd.ProjectStatusSomeday,
			Due:         new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)),
		})
		require.NoError(t, err)

		assert.Equal(t, gtd.ProjectStatusSomeday, got.Status)
		require.NotNil(t, got.Due)
		assert.True(t, got.Due.Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))

		fetched, err := db.GetProject(c, got.ID)
		require.NoError(t, err)
		assert.Equal(t, got.Outcome, fetched.Outcome)
		assert.Equal(t, got.Description, fetched.Description)
		require.NotNil(t, fetched.Due)
	})

	t.Run("default open status", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		got, err := db.CreateProject(c, gtd.Project{Title: "No explicit status"})
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusOpen, got.Status)

		fetched, err := db.GetProject(c, got.ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusOpen, fetched.Status)
	})

	t.Run("empty title rejected", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.CreateProject(c, gtd.Project{Status: gtd.ProjectStatusOpen})
		require.Error(t, err)
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.CreateProject(c, gtd.Project{Title: "Bad status", Status: "bogus"})
		require.Error(t, err)
	})
}

func TestDB_GetProject(t *testing.T) {
	t.Run("non-existent ID returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.GetProject(c, 999)
		assert.Error(t, err)
	})
}

func TestDB_ListProjects(t *testing.T) {
	t.Run("open projects sort by order_key before non-open", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.CreateProject(c, gtd.Project{Title: "Beta", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		_, err = db.CreateProject(c, gtd.Project{Title: "Alpha", Status: gtd.ProjectStatusSomeday})
		require.NoError(t, err)

		got, err := db.ListProjects(c, gtd.ProjectFilter{})
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "Beta", got[0].Title)
		assert.Equal(t, "Alpha", got[1].Title)
	})

	t.Run("open projects preserve creation order", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.CreateProject(c, gtd.Project{Title: "First"})
		require.NoError(t, err)
		_, err = db.CreateProject(c, gtd.Project{Title: "Second"})
		require.NoError(t, err)
		_, err = db.CreateProject(c, gtd.Project{Title: "Third"})
		require.NoError(t, err)

		got, err := db.ListProjects(c, gtd.ProjectFilter{})
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, "First", got[0].Title)
		assert.Equal(t, "Second", got[1].Title)
		assert.Equal(t, "Third", got[2].Title)
	})

	t.Run("filter by status", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.CreateProject(c, gtd.Project{Title: "Open", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		_, err = db.CreateProject(c, gtd.Project{Title: "Parked", Status: gtd.ProjectStatusSomeday})
		require.NoError(t, err)

		got, err := db.ListProjects(c, gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusSomeday))
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "Parked", got[0].Title)
	})
}

func TestDB_UpdateProject(t *testing.T) {
	t.Run("refreshes UpdatedAt", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateProject(c, gtd.Project{Title: "Original", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)

		created.Title = "Renamed"
		updated, err := db.UpdateProject(c, created)
		require.NoError(t, err)
		assert.False(t, updated.UpdatedAt.Before(created.UpdatedAt))

		fetched, err := db.GetProject(c, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "Renamed", fetched.Title)
	})

	t.Run("status change rejected", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateProject(c, gtd.Project{Title: "Project", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)

		created.Status = gtd.ProjectStatusDone
		_, err = db.UpdateProject(c, created)
		require.Error(t, err)
	})
}

// --- 7. Status Transitions ---

func createProjectWithTasks(t *testing.T, db *sqlite.DB, c context.Context, statuses ...gtd.TaskStatus) (gtd.Project, []gtd.Task) {
	t.Helper()
	project, err := db.CreateProject(c, gtd.Project{Title: "Test project", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	var tasks []gtd.Task
	for i, s := range statuses {
		task, err := db.CreateTask(c, gtd.Task{
			Title:     "Task " + string(rune('A'+i)),
			Status:    s,
			ProjectID: &project.ID,
		})
		require.NoError(t, err)
		tasks = append(tasks, task)
	}
	return project, tasks
}

func TestDB_CompleteProject(t *testing.T) {
	t.Run("cascade marks pending tasks done", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen, gtd.TaskStatusOpen)

		at := time.Now()
		done, err := db.CompleteProject(c, project.ID, true, at)
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusDone, done.Status)

		for _, task := range tasks {
			fetched, err := db.GetTask(c, task.ID)
			require.NoError(t, err)
			assert.Equal(t, gtd.TaskStatusDone, fetched.Status)
		}
	})

	t.Run("detach sets ProjectID to nil", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen, gtd.TaskStatusOpen)

		done, err := db.CompleteProject(c, project.ID, false, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusDone, done.Status)

		for _, task := range tasks {
			fetched, err := db.GetTask(c, task.ID)
			require.NoError(t, err)
			assert.Nil(t, fetched.ProjectID)
			assert.Equal(t, gtd.TaskStatusOpen, fetched.Status)
		}
	})

	t.Run("preserves done and dropped tasks on project", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusDone, gtd.TaskStatusDropped, gtd.TaskStatusOpen)

		_, err := db.CompleteProject(c, project.ID, true, time.Now())
		require.NoError(t, err)

		// Done task stays attached, status unchanged
		fetched, err := db.GetTask(c, tasks[0].ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusDone, fetched.Status)
		require.NotNil(t, fetched.ProjectID)
		assert.Equal(t, project.ID, *fetched.ProjectID)

		// Dropped task stays attached, status unchanged
		fetched, err = db.GetTask(c, tasks[1].ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusDropped, fetched.Status)
		require.NotNil(t, fetched.ProjectID)
		assert.Equal(t, project.ID, *fetched.ProjectID)

		// Pending task was cascaded to done
		fetched, err = db.GetTask(c, tasks[2].ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusDone, fetched.Status)
	})
}

func TestDB_DropProject(t *testing.T) {
	t.Run("cascade marks pending tasks dropped", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen, gtd.TaskStatusOpen)

		dropped, err := db.DropProject(c, project.ID, true, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusDropped, dropped.Status)

		for _, task := range tasks {
			fetched, err := db.GetTask(c, task.ID)
			require.NoError(t, err)
			assert.Equal(t, gtd.TaskStatusDropped, fetched.Status)
		}
	})

	t.Run("detach sets ProjectID to nil", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen)

		_, err := db.DropProject(c, project.ID, false, time.Now())
		require.NoError(t, err)

		fetched, err := db.GetTask(c, tasks[0].ID)
		require.NoError(t, err)
		assert.Nil(t, fetched.ProjectID)
		assert.Equal(t, gtd.TaskStatusOpen, fetched.Status)
	})
}

func TestDB_ParkProject(t *testing.T) {
	t.Run("sets status to someday without changing tasks", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen)

		parked, err := db.ParkProject(c, project.ID, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusSomeday, parked.Status)

		fetched, err := db.GetTask(c, tasks[0].ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusOpen, fetched.Status)
		require.NotNil(t, fetched.ProjectID)
		assert.Equal(t, project.ID, *fetched.ProjectID)
	})
}

func TestDB_ReopenProject(t *testing.T) {
	t.Run("reopens from someday", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, _ := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen)
		_, err := db.ParkProject(c, project.ID, time.Now())
		require.NoError(t, err)

		reopened, err := db.ReopenProject(c, project.ID, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusOpen, reopened.Status)
	})

	t.Run("reopens from done", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, _ := createProjectWithTasks(t, db, c)
		_, err := db.CompleteProject(c, project.ID, true, time.Now())
		require.NoError(t, err)

		reopened, err := db.ReopenProject(c, project.ID, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusOpen, reopened.Status)
	})

	t.Run("reopens from dropped", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, _ := createProjectWithTasks(t, db, c)
		_, err := db.DropProject(c, project.ID, true, time.Now())
		require.NoError(t, err)

		reopened, err := db.ReopenProject(c, project.ID, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.ProjectStatusOpen, reopened.Status)
	})

	t.Run("does not change task statuses", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen, gtd.TaskStatusOpen)

		// Complete with cascade → tasks become done
		_, err := db.CompleteProject(c, project.ID, true, time.Now())
		require.NoError(t, err)

		// Reopen → project open again, but tasks stay done
		_, err = db.ReopenProject(c, project.ID, time.Now())
		require.NoError(t, err)

		for _, task := range tasks {
			fetched, err := db.GetTask(c, task.ID)
			require.NoError(t, err)
			assert.Equal(t, gtd.TaskStatusDone, fetched.Status)
		}
	})
}

func TestDB_ProjectStatusChangedAt(t *testing.T) {
	t.Run("CreateProject seeds StatusChangedAt to CreatedAt", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateProject(c, gtd.Project{Title: "New", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		assert.Equal(t, created.CreatedAt, created.StatusChangedAt)

		fetched, err := db.GetProject(c, created.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, created.CreatedAt, fetched.StatusChangedAt, time.Second)
	})

	t.Run("transitions stamp StatusChangedAt with supplied instant", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, _ := createProjectWithTasks(t, db, c)
		backdate := time.Now().Add(-72 * time.Hour)

		// Park
		parked, err := db.ParkProject(c, project.ID, backdate)
		require.NoError(t, err)
		assert.WithinDuration(t, backdate.UTC(), parked.StatusChangedAt, time.Second)

		fetched, err := db.GetProject(c, project.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, backdate.UTC(), fetched.StatusChangedAt, time.Second)

		// Reopen
		later := backdate.Add(24 * time.Hour)
		reopened, err := db.ReopenProject(c, project.ID, later)
		require.NoError(t, err)
		assert.WithinDuration(t, later.UTC(), reopened.StatusChangedAt, time.Second)

		// Complete
		even_later := later.Add(24 * time.Hour)
		done, err := db.CompleteProject(c, project.ID, true, even_later)
		require.NoError(t, err)
		assert.WithinDuration(t, even_later.UTC(), done.StatusChangedAt, time.Second)

		// Reopen again, then Drop
		_, err = db.ReopenProject(c, project.ID, even_later)
		require.NoError(t, err)
		dropTime := even_later.Add(24 * time.Hour)
		dropped, err := db.DropProject(c, project.ID, true, dropTime)
		require.NoError(t, err)
		assert.WithinDuration(t, dropTime.UTC(), dropped.StatusChangedAt, time.Second)
	})

	t.Run("UpdateProject leaves StatusChangedAt unchanged", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateProject(c, gtd.Project{Title: "Original", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		orig := created.StatusChangedAt

		created.Title = "Renamed"
		_, err = db.UpdateProject(c, created)
		require.NoError(t, err)

		fetched, err := db.GetProject(c, created.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, orig, fetched.StatusChangedAt, time.Second)
	})

	t.Run("cascade stamps each task's StatusChangedAt", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, tasks := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen, gtd.TaskStatusOpen)

		backdate := time.Now().Add(-48 * time.Hour)
		_, err := db.CompleteProject(c, project.ID, true, backdate)
		require.NoError(t, err)

		for _, task := range tasks {
			fetched, err := db.GetTask(c, task.ID)
			require.NoError(t, err)
			assert.WithinDuration(t, backdate.UTC(), fetched.StatusChangedAt, time.Second)
		}
	})
}

// --- 8. Task-Project Relationship ---

func TestDB_TaskProjectRelationship(t *testing.T) {
	t.Run("task with ProjectID references valid project", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, err := db.CreateProject(c, gtd.Project{Title: "My project", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)

		task, err := db.CreateTask(c, gtd.Task{
			Title:     "Linked task",
			Status:    gtd.TaskStatusOpen,
			ProjectID: &project.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, task.ProjectID)
		assert.Equal(t, project.ID, *task.ProjectID)

		fetched, err := db.GetTask(c, task.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched.ProjectID)
		assert.Equal(t, project.ID, *fetched.ProjectID)
	})

	t.Run("default excludes someday project tasks", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, _ := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen)

		// Standalone task
		_, err := db.CreateTask(c, gtd.Task{Title: "Standalone", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		// Before parking: default filter shows both
		all, err := db.ListTasks(c, gtd.TaskFilter{})
		require.NoError(t, err)
		assert.Len(t, all, 2)

		// Park the project
		_, err = db.ParkProject(c, project.ID, time.Now())
		require.NoError(t, err)

		// After parking: default filter excludes parked project task
		filtered, err := db.ListTasks(c, gtd.TaskFilter{})
		require.NoError(t, err)
		require.Len(t, filtered, 1)
		assert.Equal(t, "Standalone", filtered[0].Title)

		// IncludeSomedayProjects: both visible again
		included, err := db.ListTasks(c, gtd.TaskFilter{IncludeSomedayProjects: true})
		require.NoError(t, err)
		assert.Len(t, included, 2)
	})

	t.Run("ProjectID filter returns only project tasks", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		project, _ := createProjectWithTasks(t, db, c, gtd.TaskStatusOpen, gtd.TaskStatusOpen)
		_, err := db.CreateTask(c, gtd.Task{Title: "Standalone", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		got, err := db.ListTasks(c, gtd.TaskFilter{}.WithProjectID(project.ID))
		require.NoError(t, err)
		assert.Len(t, got, 2)
		for _, task := range got {
			require.NotNil(t, task.ProjectID)
			assert.Equal(t, project.ID, *task.ProjectID)
		}
	})

	t.Run("pending task under closed project is visible", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		// Create project, complete with detach, then manually re-attach a
		// pending task to simulate the invariant-violation scenario.
		project, err := db.CreateProject(c, gtd.Project{Title: "Closed", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		_, err = db.CompleteProject(c, project.ID, true, time.Now())
		require.NoError(t, err)

		// Create a pending task linked to the now-closed project
		task, err := db.CreateTask(c, gtd.Task{
			Title:     "Orphaned pending",
			Status:    gtd.TaskStatusOpen,
			ProjectID: &project.ID,
		})
		require.NoError(t, err)

		// Default filter excludes someday, not done — this must be visible
		got, err := db.ListTasks(c, gtd.TaskFilter{})
		require.NoError(t, err)
		var found bool
		for _, g := range got {
			if g.ID == task.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "pending task under closed project must be visible")
	})
}

// --- Project Ordering ---

func projectTitles(t *testing.T, db *sqlite.DB, c context.Context) []string {
	t.Helper()
	got, err := db.ListProjects(c, gtd.ProjectFilter{})
	require.NoError(t, err)
	var titles []string
	for _, p := range got {
		titles = append(titles, p.Title)
	}
	return titles
}

func TestDB_MoveProjectUp(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	_, err := db.CreateProject(c, gtd.Project{Title: "A"})
	require.NoError(t, err)
	b, err := db.CreateProject(c, gtd.Project{Title: "B"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "C"})
	require.NoError(t, err)

	require.NoError(t, db.MoveProjectUp(c, b.ID, gtd.ProjectFilter{}))
	assert.Equal(t, []string{"B", "A", "C"}, projectTitles(t, db, c))
}

func TestDB_MoveProjectDown(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	_, err := db.CreateProject(c, gtd.Project{Title: "A"})
	require.NoError(t, err)
	b, err := db.CreateProject(c, gtd.Project{Title: "B"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "C"})
	require.NoError(t, err)

	require.NoError(t, db.MoveProjectDown(c, b.ID, gtd.ProjectFilter{}))
	assert.Equal(t, []string{"A", "C", "B"}, projectTitles(t, db, c))
}

func TestDB_MoveProjectUp_RejectsDoneDropped(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	p1, err := db.CreateProject(c, gtd.Project{Title: "Done"})
	require.NoError(t, err)
	_, err = db.CompleteProject(c, p1.ID, true, time.Now())
	require.NoError(t, err)

	err = db.MoveProjectUp(c, p1.ID, gtd.ProjectFilter{})
	assert.Error(t, err)

	p2, err := db.CreateProject(c, gtd.Project{Title: "Dropped"})
	require.NoError(t, err)
	_, err = db.DropProject(c, p2.ID, true, time.Now())
	require.NoError(t, err)

	err = db.MoveProjectUp(c, p2.ID, gtd.ProjectFilter{})
	assert.Error(t, err)
}

func TestDB_MoveProjectDown_WithSearchFilter(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	// Five open projects, alternating between matching "red" filter and not.
	a, err := db.CreateProject(c, gtd.Project{Title: "red A"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "blue X"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "red B"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "blue Y"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "red C"})
	require.NoError(t, err)

	filter := gtd.ProjectFilter{Search: []string{"red"}}
	require.NoError(t, db.MoveProjectDown(c, a.ID, filter))

	// Filtered view shifts A one slot.
	got, err := db.ListProjects(c, filter)
	require.NoError(t, err)
	gotTitles := make([]string, len(got))
	for i, p := range got {
		gotTitles[i] = p.Title
	}
	assert.Equal(t, []string{"red B", "red A", "red C"}, gotTitles)

	// Unfiltered: blue X / blue Y keep their keys; A lands between B and C.
	assert.Equal(t, []string{"blue X", "red B", "red A", "blue Y", "red C"}, projectTitles(t, db, c))
}

func TestDB_MoveProject_SomedayOrdersIndependently(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	// Create someday projects
	a, err := db.CreateProject(c, gtd.Project{Title: "S-A", Status: gtd.ProjectStatusSomeday})
	require.NoError(t, err)
	b, err := db.CreateProject(c, gtd.Project{Title: "S-B", Status: gtd.ProjectStatusSomeday})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "S-C", Status: gtd.ProjectStatusSomeday})
	require.NoError(t, err)

	// Create an open project — should not be affected by someday reordering
	_, err = db.CreateProject(c, gtd.Project{Title: "Open1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	// Move S-B up → S-B, S-A, S-C
	require.NoError(t, db.MoveProjectUp(c, b.ID, gtd.ProjectFilter{}))

	titles := projectTitles(t, db, c)
	// Open first, then someday
	assert.Equal(t, []string{"Open1", "S-B", "S-A", "S-C"}, titles)

	// Move S-A down → S-B, S-C, S-A
	require.NoError(t, db.MoveProjectDown(c, a.ID, gtd.ProjectFilter{}))
	titles = projectTitles(t, db, c)
	assert.Equal(t, []string{"Open1", "S-B", "S-C", "S-A"}, titles)
}

func TestDB_MoveProjectFirstLast(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a, err := db.CreateProject(c, gtd.Project{Title: "A"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "B"})
	require.NoError(t, err)
	cProj, err := db.CreateProject(c, gtd.Project{Title: "C"})
	require.NoError(t, err)

	// Initial: [A, B, C]. MoveLast(A) → [B, C, A].
	require.NoError(t, db.MoveProjectLast(c, a.ID, gtd.ProjectFilter{}))
	assert.Equal(t, []string{"B", "C", "A"}, projectTitles(t, db, c))

	// MoveFirst(C) → [C, B, A].
	require.NoError(t, db.MoveProjectFirst(c, cProj.ID, gtd.ProjectFilter{}))
	assert.Equal(t, []string{"C", "B", "A"}, projectTitles(t, db, c))

	// MoveFirst at the top and MoveLast at the bottom are silent no-ops.
	require.NoError(t, db.MoveProjectFirst(c, cProj.ID, gtd.ProjectFilter{}))
	require.NoError(t, db.MoveProjectLast(c, a.ID, gtd.ProjectFilter{}))
	assert.Equal(t, []string{"C", "B", "A"}, projectTitles(t, db, c))
}

func TestDB_MoveProjectFirst_RejectsDoneDropped(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	p1, err := db.CreateProject(c, gtd.Project{Title: "Done"})
	require.NoError(t, err)
	_, err = db.CompleteProject(c, p1.ID, true, time.Now())
	require.NoError(t, err)
	assert.Error(t, db.MoveProjectFirst(c, p1.ID, gtd.ProjectFilter{}))
	assert.Error(t, db.MoveProjectLast(c, p1.ID, gtd.ProjectFilter{}))

	p2, err := db.CreateProject(c, gtd.Project{Title: "Dropped"})
	require.NoError(t, err)
	_, err = db.DropProject(c, p2.ID, true, time.Now())
	require.NoError(t, err)
	assert.Error(t, db.MoveProjectFirst(c, p2.ID, gtd.ProjectFilter{}))
	assert.Error(t, db.MoveProjectLast(c, p2.ID, gtd.ProjectFilter{}))
}

func TestDB_MoveProjectLast_WithSearchFilter(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a, err := db.CreateProject(c, gtd.Project{Title: "red A"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "blue X"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "red B"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "blue Y"})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "red C"})
	require.NoError(t, err)

	filter := gtd.ProjectFilter{Search: []string{"red"}}
	require.NoError(t, db.MoveProjectLast(c, a.ID, filter))

	got, err := db.ListProjects(c, filter)
	require.NoError(t, err)
	gotTitles := make([]string, len(got))
	for i, p := range got {
		gotTitles[i] = p.Title
	}
	assert.Equal(t, []string{"red B", "red C", "red A"}, gotTitles)

	// Unfiltered: blue projects keep their keys; A lands after red C.
	assert.Equal(t, []string{"blue X", "red B", "blue Y", "red C", "red A"}, projectTitles(t, db, c))
}

func TestDB_MoveProjectFirst_RenumbersWhenKeysExhausted(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a, err := db.CreateProject(c, gtd.Project{Title: "A"})
	require.NoError(t, err)
	b, err := db.CreateProject(c, gtd.Project{Title: "B"})
	require.NoError(t, err)
	cProj, err := db.CreateProject(c, gtd.Project{Title: "C"})
	require.NoError(t, err)

	// Adjacent keys at the front leave no room to slot C ahead of A.
	require.NoError(t, db.SetProjectOrderKeyForTest(c, a.ID, "0"))
	require.NoError(t, db.SetProjectOrderKeyForTest(c, b.ID, "00"))
	require.NoError(t, db.SetProjectOrderKeyForTest(c, cProj.ID, "1"))

	require.NoError(t, db.MoveProjectFirst(c, cProj.ID, gtd.ProjectFilter{}))
	assert.Equal(t, []string{"C", "A", "B"}, projectTitles(t, db, c))
}

func TestDB_ListProjects_ThreeTierOrdering(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	// Create projects in various statuses
	_, err := db.CreateProject(c, gtd.Project{Title: "Open-B", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "Open-A", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "Someday-B", Status: gtd.ProjectStatusSomeday})
	require.NoError(t, err)
	_, err = db.CreateProject(c, gtd.Project{Title: "Someday-A", Status: gtd.ProjectStatusSomeday})
	require.NoError(t, err)

	// Create done/dropped with different status_changed_at times
	d1, err := db.CreateProject(c, gtd.Project{Title: "Done-Old", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	d2, err := db.CreateProject(c, gtd.Project{Title: "Dropped-New", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	earlier := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	later := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	_, err = db.CompleteProject(c, d1.ID, true, earlier)
	require.NoError(t, err)
	_, err = db.DropProject(c, d2.ID, true, later)
	require.NoError(t, err)

	titles := projectTitles(t, db, c)
	assert.Equal(t, []string{
		"Open-B", "Open-A",
		"Someday-B", "Someday-A",
		"Dropped-New", "Done-Old",
	}, titles)
}

func TestDB_ParkProject_AssignsOrderKey(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	// Create two someday projects first
	_, err := db.CreateProject(c, gtd.Project{Title: "Existing-Someday", Status: gtd.ProjectStatusSomeday})
	require.NoError(t, err)

	// Create open, then park it → should appear after existing someday
	p, err := db.CreateProject(c, gtd.Project{Title: "Newly-Parked", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	_, err = db.ParkProject(c, p.ID, time.Now())
	require.NoError(t, err)

	got, err := db.ListProjects(c, gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusSomeday))
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Existing-Someday", got[0].Title)
	assert.Equal(t, "Newly-Parked", got[1].Title)
}
