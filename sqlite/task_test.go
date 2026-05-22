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
				{Title: "Beta", Status: gtd.TaskStatusInbox},
			},
			filter: gtd.TaskFilter{},
			want:   []string{"Alpha", "Beta"},
		},
		{
			name: "closed tasks sort by updated_at desc",
			seed: []gtd.Task{
				{Title: "Older", Status: gtd.TaskStatusDone},
				{Title: "Newer", Status: gtd.TaskStatusDone},
			},
			filter: gtd.TaskFilter{}.Status(gtd.TaskStatusDone),
			want:   []string{"Newer", "Older"},
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

func TestDB_MoveUp(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusActive})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusActive})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusActive})
	// Different status should be unaffected.
	other := mustCreateTask(t, db, gtd.Task{Title: "X", Status: gtd.TaskStatusInbox})

	// Initial: [A, B, C]. MoveUp(C) → [A, C, B].
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{a.ID, cTask.ID, b.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))

	// MoveUp(C) again → [C, A, B].
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))

	// MoveUp at the top is a silent no-op.
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))

	// Other-status tasks are untouched.
	assert.Equal(t, []int64{other.ID}, taskIDs(t, db, c, gtd.TaskStatusInbox))
}

func TestDB_MoveDown(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusActive})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusActive})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusActive})

	// Initial: [A, B, C]. MoveDown(A) → [B, A, C].
	require.NoError(t, db.MoveDown(c, a.ID))
	assert.Equal(t, []int64{b.ID, a.ID, cTask.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))

	// MoveDown(A) again → [B, C, A].
	require.NoError(t, db.MoveDown(c, a.ID))
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))

	// MoveDown at the bottom is a silent no-op.
	require.NoError(t, db.MoveDown(c, a.ID))
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))
}

func TestDB_MoveUp_RejectsClosedTask(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusDone})

	assert.Error(t, db.MoveUp(c, a.ID))
	assert.Error(t, db.MoveDown(c, a.ID))
}

func TestDB_MoveUp_RenumbersWhenKeysExhausted(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusActive})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusActive})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusActive})

	// Force the first two keys to the alphabet floor so MoveUp on C
	// (landing it at position 0) cannot find space via Between and must
	// fall back to renumbering the whole status group.
	require.NoError(t, db.SetOrderKeyForTest(c, a.ID, "0"))
	require.NoError(t, db.SetOrderKeyForTest(c, b.ID, "00"))
	require.NoError(t, db.SetOrderKeyForTest(c, cTask.ID, "01"))

	// Move C up from position 2 to position 1.
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{a.ID, cTask.ID, b.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))

	// And up again — this is the move that previously errored.
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))
}

func TestDB_UpdateTask_StatusChangeAppendsToNewStatus(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusActive})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusActive})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusInbox})

	// Move C from Inbox to Active — should land at end of Active.
	cTask.Status = gtd.TaskStatusActive
	_, err := db.UpdateTask(c, cTask)
	require.NoError(t, err)

	assert.Equal(t, []int64{a.ID, b.ID, cTask.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))
}

func TestDB_DropTask_SortsByUpdatedAtDesc(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusDropped})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusActive})

	_, err := db.DropTask(c, b.ID)
	require.NoError(t, err)

	// b was just dropped, so it sorts ahead of the older a.
	assert.Equal(t, []int64{b.ID, a.ID}, taskIDs(t, db, c, gtd.TaskStatusDropped))
}

func TestDB_CreateTask_AppendsWithinStatus(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusActive})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusActive})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusActive})

	assert.Equal(t, []int64{a.ID, b.ID, cTask.ID}, taskIDs(t, db, c, gtd.TaskStatusActive))
}

func mustCreateTask(t *testing.T, db *sqlite.DB, task gtd.Task) gtd.Task {
	t.Helper()
	created, err := db.CreateTask(context.Background(), task)
	require.NoError(t, err)
	return created
}

func taskIDs(t *testing.T, db *sqlite.DB, c context.Context, status gtd.TaskStatus) []int64 {
	t.Helper()
	tasks, err := db.Tasks(c, gtd.TaskFilter{}.Status(status))
	require.NoError(t, err)
	ids := make([]int64, len(tasks))
	for i, task := range tasks {
		ids[i] = task.ID
	}
	return ids
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
