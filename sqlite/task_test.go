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
			input: gtd.Task{Title: "Buy milk", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
		},
		{
			name: "full task",
			input: gtd.Task{
				Title:       "Write proposal",
				Description: "Draft the Q3 proposal doc",
				Kind:        gtd.TaskKindNextAction,
				Status:      gtd.TaskStatusPending,
				Due:         new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)),
				DeferUntil:  new(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name:  "delegated task with assignee",
			input: gtd.Task{Title: "Review PR", Kind: gtd.TaskKindDelegated, Status: gtd.TaskStatusPending, Assignee: "alice"},
		},
		{
			name:    "missing title",
			input:   gtd.Task{Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
			wantErr: true,
		},
		{
			name:    "invalid kind",
			input:   gtd.Task{Title: "Buy milk", Kind: "bogus", Status: gtd.TaskStatusPending},
			wantErr: true,
		},
		{
			name:    "invalid status",
			input:   gtd.Task{Title: "Buy milk", Kind: gtd.TaskKindNextAction, Status: "bogus"},
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
			assert.Equal(t, tt.input.Kind, got.Kind)
			assert.Equal(t, tt.input.Status, got.Status)
			assert.Equal(t, tt.input.Assignee, got.Assignee)

			fetched, err := db.GetTask(c, got.ID)
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
			setup: gtd.Task{Title: "Old title", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
			update: func(task gtd.Task) gtd.Task {
				task.Title = "New title"
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				assert.Equal(t, "New title", task.Title)
			},
		},
		{
			name:  "set due date",
			setup: gtd.Task{Title: "Task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
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
			setup: gtd.Task{Title: "Task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending, Due: new(time.Now())},
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
			setup: gtd.Task{Title: "Task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
			update: func(task gtd.Task) gtd.Task {
				task.DeferUntil = new(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				require.NotNil(t, task.DeferUntil)
				assert.True(t, task.DeferUntil.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)))
			},
		},
		{
			name:  "change kind to delegated",
			setup: gtd.Task{Title: "Task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
			update: func(task gtd.Task) gtd.Task {
				task.Kind = gtd.TaskKindDelegated
				task.Assignee = "bob"
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				assert.Equal(t, gtd.TaskKindDelegated, task.Kind)
				assert.Equal(t, "bob", task.Assignee)
			},
		},
		{
			name:  "status change rejected",
			setup: gtd.Task{Title: "Task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
			update: func(task gtd.Task) gtd.Task {
				task.Status = gtd.TaskStatusDone
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				// error is expected — checked in the test body
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			c := ctx(t)

			created, err := db.CreateTask(c, tt.setup)
			require.NoError(t, err)

			updated := tt.update(created)
			_, err = db.UpdateTask(c, updated)

			if tt.name == "status change rejected" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			fetched, err := db.GetTask(c, created.ID)
			require.NoError(t, err)
			tt.check(t, fetched)
		})
	}
}

func TestDB_CompleteTask(t *testing.T) {
	t.Run("completes a pending task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "To complete", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
		require.NoError(t, err)

		done, err := db.CompleteTask(c, created.ID)
		require.NoError(t, err)

		assert.Equal(t, gtd.TaskStatusDone, done.Status)
		assert.False(t, done.UpdatedAt.Before(created.UpdatedAt))

		fetched, err := db.GetTask(c, created.ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusDone, fetched.Status)
	})

	t.Run("non-pending task returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Already done", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDone})
		require.NoError(t, err)

		_, err = db.CompleteTask(c, created.ID)
		assert.Error(t, err)
	})
}

func TestDB_DropTask(t *testing.T) {
	t.Run("drops a pending task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "To drop", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
		require.NoError(t, err)

		dropped, err := db.DropTask(c, created.ID)
		require.NoError(t, err)

		assert.Equal(t, gtd.TaskStatusDropped, dropped.Status)
		assert.False(t, dropped.UpdatedAt.Before(created.UpdatedAt))

		fetched, err := db.GetTask(c, created.ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusDropped, fetched.Status)
	})

	t.Run("non-pending task returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Already done", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDone})
		require.NoError(t, err)

		_, err = db.DropTask(c, created.ID)
		assert.Error(t, err)
	})

	t.Run("missing task returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.DropTask(c, 999)
		assert.Error(t, err)
	})
}

func TestDB_ReopenTask(t *testing.T) {
	t.Run("reopens a done task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Done task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDone})
		require.NoError(t, err)

		reopened, err := db.ReopenTask(c, created.ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusPending, reopened.Status)
	})

	t.Run("reopens a dropped task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Dropped task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDropped})
		require.NoError(t, err)

		reopened, err := db.ReopenTask(c, created.ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusPending, reopened.Status)
	})

	t.Run("pending task returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Pending task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
		require.NoError(t, err)

		_, err = db.ReopenTask(c, created.ID)
		assert.Error(t, err)
	})
}

func TestDB_DeleteTask(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	created, err := db.CreateTask(c, gtd.Task{Title: "To delete", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	require.NoError(t, err)

	require.NoError(t, db.DeleteTask(c, created.ID))

	_, err = db.GetTask(c, created.ID)
	assert.Error(t, err)
}

func TestDB_ListTasks(t *testing.T) {
	tests := []struct {
		name   string
		seed   []gtd.Task
		filter gtd.TaskFilter
		want   []string
	}{
		{
			name: "all tasks",
			seed: []gtd.Task{
				{Title: "Alpha", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
				{Title: "Beta", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
			},
			filter: gtd.TaskFilter{},
			want:   []string{"Alpha", "Beta"},
		},
		{
			name: "closed tasks sort by updated_at desc",
			seed: []gtd.Task{
				{Title: "Older", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDone},
				{Title: "Newer", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDone},
			},
			filter: gtd.TaskFilter{}.WithStatus(gtd.TaskStatusDone),
			want:   []string{"Newer", "Older"},
		},
		{
			name: "filter by status",
			seed: []gtd.Task{
				{Title: "Pending task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
				{Title: "Done task", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDone},
			},
			filter: gtd.TaskFilter{}.WithStatus(gtd.TaskStatusPending),
			want:   []string{"Pending task"},
		},
		{
			name: "filter by kind",
			seed: []gtd.Task{
				{Title: "Next action", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending},
				{Title: "Delegated", Kind: gtd.TaskKindDelegated, Status: gtd.TaskStatusPending, Assignee: "alice"},
			},
			filter: gtd.TaskFilter{}.WithKind(gtd.TaskKindDelegated),
			want:   []string{"Delegated"},
		},
		{
			name:   "empty result",
			seed:   []gtd.Task{{Title: "Alpha", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending}},
			filter: gtd.TaskFilter{}.WithStatus(gtd.TaskStatusDone),
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

			got, err := db.ListTasks(c, tt.filter)
			require.NoError(t, err)

			var titles []string
			for _, task := range got {
				titles = append(titles, task.Title)
			}
			assert.Equal(t, tt.want, titles)
		})
	}
}

func TestDB_ListTasks_DeferredFiltering(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	mustCreateTask(t, db, gtd.Task{Title: "Visible", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	mustCreateTask(t, db, gtd.Task{Title: "Deferred", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending, DeferUntil: &future})
	mustCreateTask(t, db, gtd.Task{Title: "Past defer", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending, DeferUntil: &past})

	// Default excludes future deferred tasks
	got, err := db.ListTasks(c, gtd.TaskFilter{})
	require.NoError(t, err)
	var titles []string
	for _, t := range got {
		titles = append(titles, t.Title)
	}
	assert.NotContains(t, titles, "Deferred")
	assert.Contains(t, titles, "Visible")
	assert.Contains(t, titles, "Past defer")

	// IncludeDeferred shows everything
	all, err := db.ListTasks(c, gtd.TaskFilter{IncludeDeferred: true})
	require.NoError(t, err)
	var allTitles []string
	for _, t := range all {
		allTitles = append(allTitles, t.Title)
	}
	assert.Contains(t, allTitles, "Deferred")
}

func TestDB_MoveUp(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})

	// Initial: [A, B, C]. MoveUp(C) → [A, C, B].
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{a.ID, cTask.ID, b.ID}, pendingIDs(t, db, c))

	// MoveUp(C) again → [C, A, B].
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, pendingIDs(t, db, c))

	// MoveUp at the top is a silent no-op.
	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveDown(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})

	// Initial: [A, B, C]. MoveDown(A) → [B, A, C].
	require.NoError(t, db.MoveDown(c, a.ID))
	assert.Equal(t, []int64{b.ID, a.ID, cTask.ID}, pendingIDs(t, db, c))

	// MoveDown(A) again → [B, C, A].
	require.NoError(t, db.MoveDown(c, a.ID))
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, pendingIDs(t, db, c))

	// MoveDown at the bottom is a silent no-op.
	require.NoError(t, db.MoveDown(c, a.ID))
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveUp_RejectsClosedTask(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDone})

	assert.Error(t, db.MoveUp(c, a.ID))
	assert.Error(t, db.MoveDown(c, a.ID))
}

func TestDB_MoveUp_RenumbersWhenKeysExhausted(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})

	require.NoError(t, db.SetOrderKeyForTest(c, a.ID, "0"))
	require.NoError(t, db.SetOrderKeyForTest(c, b.ID, "00"))
	require.NoError(t, db.SetOrderKeyForTest(c, cTask.ID, "01"))

	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{a.ID, cTask.ID, b.ID}, pendingIDs(t, db, c))

	require.NoError(t, db.MoveUp(c, cTask.ID))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, pendingIDs(t, db, c))
}

func TestDB_CreateTask_AppendsWithinPending(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})

	assert.Equal(t, []int64{a.ID, b.ID, cTask.ID}, pendingIDs(t, db, c))
}

func TestDB_DropTask_SortsByUpdatedAtDesc(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusDropped})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Kind: gtd.TaskKindNextAction, Status: gtd.TaskStatusPending})

	_, err := db.DropTask(c, b.ID)
	require.NoError(t, err)

	// b was just dropped, so it sorts ahead of the older a.
	assert.Equal(t, []int64{b.ID, a.ID}, taskIDsByStatus(t, db, c, gtd.TaskStatusDropped))
}

func mustCreateTask(t *testing.T, db *sqlite.DB, task gtd.Task) gtd.Task {
	t.Helper()
	created, err := db.CreateTask(context.Background(), task)
	require.NoError(t, err)
	return created
}

func pendingIDs(t *testing.T, db *sqlite.DB, c context.Context) []int64 {
	t.Helper()
	return taskIDsByStatus(t, db, c, gtd.TaskStatusPending)
}

func taskIDsByStatus(t *testing.T, db *sqlite.DB, c context.Context, status gtd.TaskStatus) []int64 {
	t.Helper()
	tasks, err := db.ListTasks(c, gtd.TaskFilter{}.WithStatus(status))
	require.NoError(t, err)
	ids := make([]int64, len(tasks))
	for i, task := range tasks {
		ids[i] = task.ID
	}
	return ids
}

