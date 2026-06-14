package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

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
			input: gtd.Task{Title: "Buy milk", Status: gtd.TaskStatusOpen},
		},
		{
			name: "full task",
			input: gtd.Task{
				Title:       "Write proposal",
				Description: "Draft the Q3 proposal doc",
				Status:      gtd.TaskStatusOpen,
				Due:         new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)),
				DeferUntil:  new(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name:  "delegated task with assignee",
			input: gtd.Task{Title: "Review PR", Status: gtd.TaskStatusOpen, Assignee: new("alice")},
		},
		{
			name:    "missing title",
			input:   gtd.Task{Status: gtd.TaskStatusOpen},
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
			setup: gtd.Task{Title: "Old title", Status: gtd.TaskStatusOpen},
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
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusOpen},
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
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusOpen, Due: new(time.Now())},
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
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusOpen},
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
			name:  "set assignee",
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusOpen},
			update: func(task gtd.Task) gtd.Task {
				task.Assignee = new("bob")
				return task
			},
			check: func(t *testing.T, task gtd.Task) {
				require.NotNil(t, task.Assignee)
				assert.Equal(t, "bob", *task.Assignee)
			},
		},
		{
			name:  "status change rejected",
			setup: gtd.Task{Title: "Task", Status: gtd.TaskStatusOpen},
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

		created, err := db.CreateTask(c, gtd.Task{Title: "To complete", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		done, err := db.CompleteTask(c, created.ID, time.Now())
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

		created, err := db.CreateTask(c, gtd.Task{Title: "Already done", Status: gtd.TaskStatusDone})
		require.NoError(t, err)

		_, err = db.CompleteTask(c, created.ID, time.Now())
		assert.Error(t, err)
	})
}

func TestDB_DropTask(t *testing.T) {
	t.Run("drops a pending task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "To drop", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		dropped, err := db.DropTask(c, created.ID, time.Now())
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

		created, err := db.CreateTask(c, gtd.Task{Title: "Already done", Status: gtd.TaskStatusDone})
		require.NoError(t, err)

		_, err = db.DropTask(c, created.ID, time.Now())
		assert.Error(t, err)
	})

	t.Run("missing task returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.DropTask(c, 999, time.Now())
		assert.Error(t, err)
	})
}

func TestDB_ReopenTask(t *testing.T) {
	t.Run("reopens a done task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Done task", Status: gtd.TaskStatusDone})
		require.NoError(t, err)

		reopened, err := db.ReopenTask(c, created.ID, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusOpen, reopened.Status)
	})

	t.Run("reopens a dropped task", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Dropped task", Status: gtd.TaskStatusDropped})
		require.NoError(t, err)

		reopened, err := db.ReopenTask(c, created.ID, time.Now())
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusOpen, reopened.Status)
	})

	t.Run("open task returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Open task", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		_, err = db.ReopenTask(c, created.ID, time.Now())
		assert.Error(t, err)
	})
}

func TestDB_StatusChangedAt(t *testing.T) {
	t.Run("CreateTask sets status_changed_at to created_at", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "New", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)
		assert.Equal(t, created.CreatedAt, created.StatusChangedAt)

		fetched, err := db.GetTask(c, created.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, created.CreatedAt, fetched.StatusChangedAt, time.Second)
	})

	t.Run("backdated transition stores the instant while updated_at advances", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "To complete", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		backdate := time.Now().Add(-72 * time.Hour)
		done, err := db.CompleteTask(c, created.ID, backdate)
		require.NoError(t, err)

		assert.WithinDuration(t, backdate.UTC(), done.StatusChangedAt, time.Second)
		assert.False(t, done.UpdatedAt.Before(created.UpdatedAt))
		assert.True(t, done.StatusChangedAt.Before(done.UpdatedAt))

		fetched, err := db.GetTask(c, created.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, backdate, fetched.StatusChangedAt, time.Second)
	})

	t.Run("non-status UpdateTask leaves status_changed_at unchanged", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		created, err := db.CreateTask(c, gtd.Task{Title: "Original", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)
		orig := created.StatusChangedAt

		created.Title = "Renamed"
		_, err = db.UpdateTask(c, created)
		require.NoError(t, err)

		fetched, err := db.GetTask(c, created.ID)
		require.NoError(t, err)
		assert.WithinDuration(t, orig, fetched.StatusChangedAt, time.Second)
	})
}

func TestMigration_StatusChangedAtBackfill(t *testing.T) {
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	m1, err := sqlite.MigrationSQL("0001_tasks.sql")
	require.NoError(t, err)
	_, err = conn.Exec(m1)
	require.NoError(t, err)

	const updated = "2026-01-02T03:04:05.000Z"
	_, err = conn.Exec(
		`INSERT INTO tasks (title, status, created_at, updated_at) VALUES ('old', 'done', ?, ?)`,
		"2026-01-01T00:00:00.000Z", updated,
	)
	require.NoError(t, err)

	m2, err := sqlite.MigrationSQL("0002_task_status_changed_at.sql")
	require.NoError(t, err)
	_, err = conn.Exec(m2)
	require.NoError(t, err)

	var gotStatusChangedAt, gotUpdatedAt string
	err = conn.QueryRow(`SELECT status_changed_at, updated_at FROM tasks`).Scan(&gotStatusChangedAt, &gotUpdatedAt)
	require.NoError(t, err)
	assert.Equal(t, gotUpdatedAt, gotStatusChangedAt)
}

func TestDB_DeleteTask(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	created, err := db.CreateTask(c, gtd.Task{Title: "To delete", Status: gtd.TaskStatusOpen})
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
				{Title: "Alpha", Status: gtd.TaskStatusOpen},
				{Title: "Beta", Status: gtd.TaskStatusOpen},
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
			filter: gtd.TaskFilter{}.WithStatus(gtd.TaskStatusDone),
			want:   []string{"Newer", "Older"},
		},
		{
			name: "filter by status",
			seed: []gtd.Task{
				{Title: "Pending task", Status: gtd.TaskStatusOpen},
				{Title: "Done task", Status: gtd.TaskStatusDone},
			},
			filter: gtd.TaskFilter{}.WithStatus(gtd.TaskStatusOpen),
			want:   []string{"Pending task"},
		},
		{
			name:   "empty result",
			seed:   []gtd.Task{{Title: "Alpha", Status: gtd.TaskStatusOpen}},
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

func TestDB_ListTasks_NoImplicitDeferralFiltering(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	mustCreateTask(t, db, gtd.Task{Title: "Visible", Status: gtd.TaskStatusOpen})
	mustCreateTask(t, db, gtd.Task{Title: "Deferred", Status: gtd.TaskStatusOpen, DeferUntil: &future})
	mustCreateTask(t, db, gtd.Task{Title: "Past defer", Status: gtd.TaskStatusOpen, DeferUntil: &past})

	// With no Ready/Defer predicate, future-deferred tasks are returned.
	got, err := db.ListTasks(c, gtd.TaskFilter{})
	require.NoError(t, err)
	var titles []string
	for _, t := range got {
		titles = append(titles, t.Title)
	}
	assert.Contains(t, titles, "Deferred")
	assert.Contains(t, titles, "Visible")
	assert.Contains(t, titles, "Past defer")
}

func TestDB_ListTasks_TextAndDateFilters(t *testing.T) {
	now := time.Now()
	future := now.Add(48 * time.Hour)
	past := now.Add(-48 * time.Hour)
	dueSoon := now.Add(12 * time.Hour)
	dueOverdue := now.Add(-12 * time.Hour)

	availableAsOfNow := &gtd.DatePredicate{Kind: gtd.AvailableAsOf, Time: now.UTC()}

	seed := func(db *sqlite.DB, c context.Context) {
		mustCreateTask(t, db, gtd.Task{Title: "Write report", Description: "quarterly numbers", Status: gtd.TaskStatusOpen, Assignee: new("Bob")})
		mustCreateTask(t, db, gtd.Task{Title: "Review report", Description: "for alice", Status: gtd.TaskStatusOpen, Assignee: new("Carol")})
		mustCreateTask(t, db, gtd.Task{Title: "Call plumber", Status: gtd.TaskStatusOpen})
		mustCreateTask(t, db, gtd.Task{Title: "Deferred future", Status: gtd.TaskStatusOpen, DeferUntil: &future})
		mustCreateTask(t, db, gtd.Task{Title: "Deferred past", Status: gtd.TaskStatusOpen, DeferUntil: &past})
		mustCreateTask(t, db, gtd.Task{Title: "Due soon", Status: gtd.TaskStatusOpen, Due: &dueSoon})
		mustCreateTask(t, db, gtd.Task{Title: "Overdue", Status: gtd.TaskStatusOpen, Due: &dueOverdue})
	}

	tests := []struct {
		name   string
		filter gtd.TaskFilter
		want   []string
	}{
		{
			name:   "free-text single term across columns",
			filter: gtd.TaskFilter{}.WithSearch("report"),
			want:   []string{"Write report", "Review report"},
		},
		{
			name:   "free-text matches description",
			filter: gtd.TaskFilter{}.WithSearch("quarterly"),
			want:   []string{"Write report"},
		},
		{
			name:   "free-text matches assignee",
			filter: gtd.TaskFilter{}.WithSearch("bob"),
			want:   []string{"Write report"},
		},
		{
			name:   "multiple terms ANDed",
			filter: gtd.TaskFilter{}.WithSearch("report", "bob"),
			want:   []string{"Write report"},
		},
		{
			name:   "assignee filter case-insensitive",
			filter: gtd.TaskFilter{}.WithAssignee("carol"),
			want:   []string{"Review report"},
		},
		{
			name:   "due cumulative includes overdue excludes null",
			filter: gtd.TaskFilter{Due: &gtd.DatePredicate{Kind: gtd.OnOrBefore, Time: now.UTC()}},
			want:   []string{"Overdue"},
		},
		{
			name:   "ready includes null and passed gates",
			filter: gtd.TaskFilter{Ready: availableAsOfNow},
			want:   []string{"Write report", "Review report", "Call plumber", "Deferred past", "Due soon", "Overdue"},
		},
		{
			name:   "defer strict lower bound excludes null",
			filter: gtd.TaskFilter{Defer: &gtd.DatePredicate{Kind: gtd.After, Time: now.UTC()}},
			want:   []string{"Deferred future"},
		},
		{
			name:   "defer IsNull",
			filter: gtd.TaskFilter{Defer: &gtd.DatePredicate{Kind: gtd.IsNull}, Search: []string{"deferred"}},
			want:   nil,
		},
		{
			name:   "defer IsNotNull",
			filter: gtd.TaskFilter{Defer: &gtd.DatePredicate{Kind: gtd.IsNotNull}},
			want:   []string{"Deferred future", "Deferred past"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			c := ctx(t)
			seed(db, c)

			got, err := db.ListTasks(c, tt.filter)
			require.NoError(t, err)

			var titles []string
			for _, task := range got {
				titles = append(titles, task.Title)
			}
			assert.ElementsMatch(t, tt.want, titles)
		})
	}
}

func TestDB_MoveUp(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusOpen})

	// Initial: [A, B, C]. MoveUp(C) → [A, C, B].
	require.NoError(t, db.MoveTaskUp(c, cTask.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{a.ID, cTask.ID, b.ID}, pendingIDs(t, db, c))

	// MoveUp(C) again → [C, A, B].
	require.NoError(t, db.MoveTaskUp(c, cTask.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, pendingIDs(t, db, c))

	// MoveUp at the top is a silent no-op.
	require.NoError(t, db.MoveTaskUp(c, cTask.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveDown(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusOpen})

	// Initial: [A, B, C]. MoveDown(A) → [B, A, C].
	require.NoError(t, db.MoveTaskDown(c, a.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{b.ID, a.ID, cTask.ID}, pendingIDs(t, db, c))

	// MoveDown(A) again → [B, C, A].
	require.NoError(t, db.MoveTaskDown(c, a.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, pendingIDs(t, db, c))

	// MoveDown at the bottom is a silent no-op.
	require.NoError(t, db.MoveTaskDown(c, a.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveUp_RejectsClosedTask(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusDone})

	assert.Error(t, db.MoveTaskUp(c, a.ID, gtd.TaskFilter{}))
	assert.Error(t, db.MoveTaskDown(c, a.ID, gtd.TaskFilter{}))
}

func TestDB_MoveUp_RenumbersWhenKeysExhausted(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusOpen})

	require.NoError(t, db.SetOrderKeyForTest(c, a.ID, "0"))
	require.NoError(t, db.SetOrderKeyForTest(c, b.ID, "00"))
	require.NoError(t, db.SetOrderKeyForTest(c, cTask.ID, "01"))

	require.NoError(t, db.MoveTaskUp(c, cTask.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{a.ID, cTask.ID, b.ID}, pendingIDs(t, db, c))

	require.NoError(t, db.MoveTaskUp(c, cTask.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveTaskDown_WithSearchFilter(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	// Tasks: red-A, blue-X, red-B, blue-Y, red-C. Filter "red" sees [A, B, C].
	// MoveDown(A, filter=red) should land A between B and C, regardless of X/Y.
	a := mustCreateTask(t, db, gtd.Task{Title: "red A", Status: gtd.TaskStatusOpen})
	x := mustCreateTask(t, db, gtd.Task{Title: "blue X", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "red B", Status: gtd.TaskStatusOpen})
	y := mustCreateTask(t, db, gtd.Task{Title: "blue Y", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "red C", Status: gtd.TaskStatusOpen})

	filter := gtd.TaskFilter{Search: []string{"red"}}
	require.NoError(t, db.MoveTaskDown(c, a.ID, filter))

	// In the filtered view, A is now between B and C: [B, A, C].
	got := listTitles(t, db, c, filter)
	assert.Equal(t, []string{"red B", "red A", "red C"}, got)

	// In an unfiltered view, X/Y keep their original keys; A's new key sits
	// between B's and C's. Full order: [X, B, A, Y, C].
	assert.Equal(t, []int64{x.ID, b.ID, a.ID, y.ID, cTask.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveTaskDown_OneItemFilterIsNoop(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	mustCreateTask(t, db, gtd.Task{Title: "blue X", Status: gtd.TaskStatusOpen})
	only := mustCreateTask(t, db, gtd.Task{Title: "red only", Status: gtd.TaskStatusOpen})
	mustCreateTask(t, db, gtd.Task{Title: "blue Y", Status: gtd.TaskStatusOpen})

	before := pendingIDs(t, db, c)
	require.NoError(t, db.MoveTaskDown(c, only.ID, gtd.TaskFilter{Search: []string{"red"}}))
	require.NoError(t, db.MoveTaskUp(c, only.ID, gtd.TaskFilter{Search: []string{"red"}}))
	assert.Equal(t, before, pendingIDs(t, db, c))
}

func TestDB_MoveTaskDown_InProjectFilterStaysInProject(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	p1, err := db.CreateProject(c, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	p2, err := db.CreateProject(c, gtd.Project{Title: "P2", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	// Interleave tasks across two projects: a(P1), x(P2), b(P1), y(P2), cT(P1).
	a := mustCreateTask(t, db, gtd.Task{Title: "a", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	x := mustCreateTask(t, db, gtd.Task{Title: "x", Status: gtd.TaskStatusOpen, ProjectID: &p2.ID})
	b := mustCreateTask(t, db, gtd.Task{Title: "b", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
	y := mustCreateTask(t, db, gtd.Task{Title: "y", Status: gtd.TaskStatusOpen, ProjectID: &p2.ID})
	cT := mustCreateTask(t, db, gtd.Task{Title: "c", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})

	// MoveDown(a) within project P1: a should land between b and c in P1.
	require.NoError(t, db.MoveTaskDown(c, a.ID, gtd.TaskFilter{ProjectID: &p1.ID}))

	// P1's view: [b, a, c]. P2's view: [x, y] unchanged.
	p1Tasks, err := db.ListTasks(c, gtd.TaskFilter{ProjectID: &p1.ID})
	require.NoError(t, err)
	assert.Equal(t, []int64{b.ID, a.ID, cT.ID}, taskIDs(p1Tasks))

	p2Tasks, err := db.ListTasks(c, gtd.TaskFilter{ProjectID: &p2.ID})
	require.NoError(t, err)
	assert.Equal(t, []int64{x.ID, y.ID}, taskIDs(p2Tasks))
}

func TestDB_MoveTaskFirstLast(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusOpen})

	// Initial: [A, B, C]. MoveLast(A) → [B, C, A].
	require.NoError(t, db.MoveTaskLast(c, a.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{b.ID, cTask.ID, a.ID}, pendingIDs(t, db, c))

	// MoveFirst(C) → [C, B, A].
	require.NoError(t, db.MoveTaskFirst(c, cTask.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{cTask.ID, b.ID, a.ID}, pendingIDs(t, db, c))

	// MoveFirst at the top and MoveLast at the bottom are silent no-ops.
	require.NoError(t, db.MoveTaskFirst(c, cTask.ID, gtd.TaskFilter{}))
	require.NoError(t, db.MoveTaskLast(c, a.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{cTask.ID, b.ID, a.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveTaskFirstLast_RejectsClosedTask(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusDone})

	assert.Error(t, db.MoveTaskFirst(c, a.ID, gtd.TaskFilter{}))
	assert.Error(t, db.MoveTaskLast(c, a.ID, gtd.TaskFilter{}))
}

func TestDB_MoveTaskLast_WithSearchFilter(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	// Tasks: red A, blue X, red B, blue Y, red C. Filter "red" sees [A, B, C].
	// MoveLast(A, filter=red) lands A after C among the filtered tasks,
	// leaving the blue tasks' keys untouched.
	a := mustCreateTask(t, db, gtd.Task{Title: "red A", Status: gtd.TaskStatusOpen})
	x := mustCreateTask(t, db, gtd.Task{Title: "blue X", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "red B", Status: gtd.TaskStatusOpen})
	y := mustCreateTask(t, db, gtd.Task{Title: "blue Y", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "red C", Status: gtd.TaskStatusOpen})

	filter := gtd.TaskFilter{Search: []string{"red"}}
	require.NoError(t, db.MoveTaskLast(c, a.ID, filter))

	assert.Equal(t, []string{"red B", "red C", "red A"}, listTitles(t, db, c, filter))

	// Unfiltered: blue X / blue Y keep their original keys; A lands after C.
	assert.Equal(t, []int64{x.ID, b.ID, y.ID, cTask.ID, a.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveTaskFirst_WithSearchFilter(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "red A", Status: gtd.TaskStatusOpen})
	x := mustCreateTask(t, db, gtd.Task{Title: "blue X", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "red B", Status: gtd.TaskStatusOpen})
	y := mustCreateTask(t, db, gtd.Task{Title: "blue Y", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "red C", Status: gtd.TaskStatusOpen})

	filter := gtd.TaskFilter{Search: []string{"red"}}
	require.NoError(t, db.MoveTaskFirst(c, cTask.ID, filter))

	assert.Equal(t, []string{"red C", "red A", "red B"}, listTitles(t, db, c, filter))

	// Unfiltered: blue X / blue Y keep their original keys; C lands before A.
	assert.Equal(t, []int64{cTask.ID, a.ID, x.ID, b.ID, y.ID}, pendingIDs(t, db, c))
}

func TestDB_MoveTaskFirst_RenumbersWhenKeysExhausted(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusOpen})

	// Adjacent keys at the front leave no room to slot C ahead of A.
	require.NoError(t, db.SetOrderKeyForTest(c, a.ID, "0"))
	require.NoError(t, db.SetOrderKeyForTest(c, b.ID, "00"))
	require.NoError(t, db.SetOrderKeyForTest(c, cTask.ID, "1"))

	require.NoError(t, db.MoveTaskFirst(c, cTask.ID, gtd.TaskFilter{}))
	assert.Equal(t, []int64{cTask.ID, a.ID, b.ID}, pendingIDs(t, db, c))
}

func listTitles(t *testing.T, db *sqlite.DB, ctx context.Context, filter gtd.TaskFilter) []string {
	t.Helper()
	tasks, err := db.ListTasks(ctx, filter)
	require.NoError(t, err)
	out := make([]string, len(tasks))
	for i, task := range tasks {
		out[i] = task.Title
	}
	return out
}

func taskIDs(tasks []gtd.Task) []int64 {
	ids := make([]int64, len(tasks))
	for i, task := range tasks {
		ids[i] = task.ID
	}
	return ids
}

func TestDB_CreateTask_AppendsWithinPending(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusOpen})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusOpen})
	cTask := mustCreateTask(t, db, gtd.Task{Title: "C", Status: gtd.TaskStatusOpen})

	assert.Equal(t, []int64{a.ID, b.ID, cTask.ID}, pendingIDs(t, db, c))
}

func TestDB_DropTask_SortsByUpdatedAtDesc(t *testing.T) {
	db := openTestDB(t)
	c := ctx(t)

	a := mustCreateTask(t, db, gtd.Task{Title: "A", Status: gtd.TaskStatusDropped})
	b := mustCreateTask(t, db, gtd.Task{Title: "B", Status: gtd.TaskStatusOpen})

	_, err := db.DropTask(c, b.ID, time.Now())
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
	return taskIDsByStatus(t, db, c, gtd.TaskStatusOpen)
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
