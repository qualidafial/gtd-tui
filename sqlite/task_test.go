package sqlite_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
)

func TestDB_CreateTask(t *testing.T) {
	tests := []struct {
		name    string
		input   gtd.Task
		wantErr bool
	}{
		{
			name:  "minimal task",
			input: gtd.Task{Title: "Buy milk", Status: gtd.TaskStatusInbox},
		},
		{
			name: "full task",
			input: gtd.Task{
				Title:       "Write proposal",
				Description: "Draft the Q3 proposal doc",
				Status:      gtd.TaskStatusActive,
				Due:         new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)),
				DeferUntil:  new(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name:    "missing title",
			input:   gtd.Task{Status: gtd.TaskStatusInbox},
			wantErr: true,
		},
		{
			name:    "missing status",
			input:   gtd.Task{Title: "Buy milk"},
			wantErr: true,
		},
		{
			name:    "invalid status",
			input:   gtd.Task{Title: "Buy milk", Status: "bogus"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			c := ctx(t)

			got, err := db.CreateTask(c, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.NotZero(t, got.ID)
			assert.False(t, got.CreatedAt.IsZero())
			assert.False(t, got.UpdatedAt.IsZero())
			assert.Equal(t, tt.input.Title, got.Title)
			assert.Equal(t, tt.input.Description, got.Description)
			assert.Equal(t, tt.input.Status, got.Status)

			fetched, err := db.Task(c, got.ID)
			require.NoError(t, err)
			assert.Equal(t, got, fetched)
		})
	}
}

func TestDB_UpdateTask(t *testing.T) {
	tests := []struct {
		name   string
		setup  gtd.Task
		update func(gtd.Task) gtd.Task
		check  func(*testing.T, gtd.Task)
	}{
		{
			name:  "update title",
			setup: gtd.Task{Title: "Old title", Status: gtd.TaskStatusInbox},
			update: func(task gtd.Task) gtd.Task {
				task.Title = "New title"
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				assert.Equal(t, "New title", task.Title)
			},
		},
		{
			name:  "update status",
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusInbox},
			update: func(task gtd.Task) gtd.Task {
				task.Status = gtd.TaskStatusActive
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				assert.Equal(t, gtd.TaskStatusActive, task.Status)
			},
		},
		{
			name:  "set due date",
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusActive},
			update: func(task gtd.Task) gtd.Task {
				task.Due = new(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				require.NotNil(t, task.Due)
				assert.True(t, task.Due.Equal(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)))
			},
		},
		{
			name:  "clear due date",
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusActive, Due: new(time.Now())},
			update: func(task gtd.Task) gtd.Task {
				task.Due = nil
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				assert.Nil(t, task.Due)
			},
		},
		{
			name:  "set defer until",
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusDeferred},
			update: func(task gtd.Task) gtd.Task {
				task.DeferUntil = new(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				require.NotNil(t, task.DeferUntil)
				assert.True(t, task.DeferUntil.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			c := ctx(t)

			created, err := db.CreateTask(c, tt.setup)
			require.NoError(t, err)

			_, err = db.UpdateTask(c, tt.update(created))
			require.NoError(t, err)

			fetched, err := db.Task(c, created.ID)
			require.NoError(t, err)
			tt.check(t, fetched)
		})
	}
}

func TestDB_DropTask(t *testing.T) {
	t.Run("drops an existing task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "To drop", Status: gtd.TaskStatusActive})
		require.NoError(t, err)

		dropped, err := db.DropTask(c, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, dropped.ID)
		assert.Equal(t, created.Title, dropped.Title)
		assert.Equal(t, gtd.TaskStatusDropped, dropped.Status)
		assert.False(t, dropped.UpdatedAt.Before(created.UpdatedAt))

		fetched, err := db.Task(c, created.ID)
		require.NoError(t, err)
		assert.Equal(t, dropped, fetched)
	})

	t.Run("missing task returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.DropTask(c, 999)
		assert.Error(t, err)
	})
}

func TestDB_DeleteTask(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	created, err := db.CreateTask(c, gtd.Task{Title: "To delete", Status: gtd.TaskStatusInbox})
	require.NoError(t, err)

	require.NoError(t, db.DeleteTask(c, created.ID))

	_, err = db.Task(c, created.ID)
	assert.Error(t, err)
}

func TestDB_Tasks(t *testing.T) {
	active := gtd.TaskStatusActive

	tests := []struct {
		name   string
		seed   []gtd.Task
		filter gtd.TaskFilter
		want   []string
	}{
		{
			name: "all tasks",
			seed: []gtd.Task{
				{Title: "Alpha", Status: gtd.TaskStatusInbox},
				{Title: "Beta", Status: gtd.TaskStatusActive},
			},
			filter: gtd.TaskFilter{},
			want:   []string{"Alpha", "Beta"},
		},
		{
			name: "filter by status",
			seed: []gtd.Task{
				{Title: "Inbox task", Status: gtd.TaskStatusInbox},
				{Title: "Active task", Status: gtd.TaskStatusActive},
			},
			filter: gtd.TaskFilter{}.Status(active),
			want:   []string{"Active task"},
		},
		{
			name:   "empty result",
			seed:   []gtd.Task{{Title: "Alpha", Status: gtd.TaskStatusInbox}},
			filter: gtd.TaskFilter{}.Status(active),
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			c := ctx(t)

			for _, task := range tt.seed {
				_, err := db.CreateTask(c, task)
				require.NoError(t, err)
			}

			got, err := db.Tasks(c, tt.filter)
			require.NoError(t, err)

			var titles []string
			for _, task := range got {
				titles = append(titles, task.Title)
			}
			assert.Equal(t, tt.want, titles)
		})
	}
}

// func TestDB_FilterTasksByProject(t *testing.T) {
// 	db := openTestDB(t)
// 	c := ctx(t)

// 	projA, err := db.CreateProject(c, gtd.Project{Title: "Project A", Status: gtd.ProjectStatusActive})
// 	require.NoError(t, err)
// 	projB, err := db.CreateProject(c, gtd.Project{Title: "Project B", Status: gtd.ProjectStatusActive})
// 	require.NoError(t, err)

// 	taskA, err := db.CreateTask(c, gtd.Task{Title: "Task A", Status: gtd.TaskStatusActive})
// 	require.NoError(t, err)
// 	require.NoError(t, db.AddTaskToProject(c, taskA.ID, projA.ID))

// 	taskB, err := db.CreateTask(c, gtd.Task{Title: "Task B", Status: gtd.TaskStatusActive})
// 	require.NoError(t, err)
// 	require.NoError(t, db.AddTaskToProject(c, taskB.ID, projB.ID))

// 	taskAB, err := db.CreateTask(c, gtd.Task{Title: "Task AB", Status: gtd.TaskStatusActive})
// 	require.NoError(t, err)
// 	require.NoError(t, db.AddTaskToProject(c, taskAB.ID, projA.ID))
// 	require.NoError(t, db.AddTaskToProject(c, taskAB.ID, projB.ID))

// 	got, err := db.Tasks(c, gtd.TaskFilter{
// 		ProjectIDs: []int64{projA.ID},
// 	})
// 	require.NoError(t, err)

// 	var ids []int64
// 	for _, task := range got {
// 		ids = append(ids, task.ID)
// 	}
// 	assert.ElementsMatch(t, []int64{taskA.ID, taskAB.ID}, ids)
// }
