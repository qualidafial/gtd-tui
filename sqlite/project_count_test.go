package sqlite_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
)

func TestDB_CountTasksByProjects(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	proj, err := db.CreateProject(c, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	mkTask := func(status gtd.TaskStatus) {
		t.Helper()
		id := proj.ID
		_, err := db.CreateTask(c, gtd.Task{Title: "t", Status: status, ProjectID: &id})
		require.NoError(t, err)
	}

	mkTask(gtd.TaskStatusOpen)
	mkTask(gtd.TaskStatusOpen)
	mkTask(gtd.TaskStatusDone)
	mkTask(gtd.TaskStatusDropped) // excluded from both counts

	counts, err := db.CountTasksByProjects(c, []int64{proj.ID})
	require.NoError(t, err)

	got := counts[proj.ID]
	assert.Equal(t, 1, got.Complete, "complete (done tasks)")
	assert.Equal(t, 3, got.Total, "total (dropped excluded)")
}

func TestDB_CountTasksByProjects_Empty(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	counts, err := db.CountTasksByProjects(c, nil)
	require.NoError(t, err)
	assert.Nil(t, counts)
}

func TestDB_CountTasksByProjects_NoTasks(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	proj, err := db.CreateProject(c, gtd.Project{Title: "empty", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	counts, err := db.CountTasksByProjects(c, []int64{proj.ID})
	require.NoError(t, err)
	// project with no tasks returns zero value (not present in map)
	assert.Equal(t, gtd.ProjectTaskCounts{}, counts[proj.ID])
}

func TestDB_CountTasksByProjects_MultipleProjects(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	p1, err := db.CreateProject(c, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	p2, err := db.CreateProject(c, gtd.Project{Title: "P2", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	mk := func(proj gtd.Project, status gtd.TaskStatus) {
		t.Helper()
		id := proj.ID
		_, err := db.CreateTask(c, gtd.Task{Title: "t", Status: status, ProjectID: &id})
		require.NoError(t, err)
	}

	mk(p1, gtd.TaskStatusOpen)
	mk(p1, gtd.TaskStatusDone)
	mk(p2, gtd.TaskStatusDropped) // all dropped → zero total

	counts, err := db.CountTasksByProjects(c, []int64{p1.ID, p2.ID})
	require.NoError(t, err)

	assert.Equal(t, gtd.ProjectTaskCounts{Complete: 1, Total: 2}, counts[p1.ID])
	assert.Equal(t, gtd.ProjectTaskCounts{}, counts[p2.ID], "all-dropped project absent from result")
}