package sqlite_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
)

func TestDB_CreateItem(t *testing.T) {
	t.Run("minimal item", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		got, err := db.CreateItem(c, gtd.Item{Title: "Call dentist"})
		require.NoError(t, err)

		assert.NotZero(t, got.ID)
		assert.False(t, got.CreatedAt.IsZero())
		assert.False(t, got.UpdatedAt.IsZero())
		assert.Equal(t, "Call dentist", got.Title)
		assert.Empty(t, got.Description)
		assert.False(t, got.Discarded)
		assert.Nil(t, got.ClarifiedIntoTaskID)
		assert.Nil(t, got.ClarifiedIntoProjectID)
	})

	t.Run("with description", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		got, err := db.CreateItem(c, gtd.Item{
			Title:       "Plan party",
			Description: "Friday after work",
		})
		require.NoError(t, err)

		fetched, err := db.GetItem(c, got.ID)
		require.NoError(t, err)
		assert.Equal(t, "Plan party", fetched.Title)
		assert.Equal(t, "Friday after work", fetched.Description)
	})

	t.Run("empty title rejected", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		_, err := db.CreateItem(c, gtd.Item{Title: ""})
		require.Error(t, err)
	})

	t.Run("UTC timestamps", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		before := time.Now().UTC().Add(-time.Second)
		got, err := db.CreateItem(c, gtd.Item{Title: "x"})
		require.NoError(t, err)
		after := time.Now().UTC().Add(time.Second)

		assert.WithinRange(t, got.CreatedAt, before, after)
		assert.Equal(t, time.UTC, got.CreatedAt.Location())
	})
}

func TestDB_ListItems(t *testing.T) {
	t.Run("returns unclarified items in FIFO order", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		a, err := db.CreateItem(c, gtd.Item{Title: "first"})
		require.NoError(t, err)
		// Ensure distinct created_at ordering.
		time.Sleep(2 * time.Millisecond)
		b, err := db.CreateItem(c, gtd.Item{Title: "second"})
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond)
		cc, err := db.CreateItem(c, gtd.Item{Title: "third"})
		require.NoError(t, err)

		items, err := db.ListItems(c)
		require.NoError(t, err)
		require.Len(t, items, 3)
		assert.Equal(t, a.ID, items[0].ID)
		assert.Equal(t, b.ID, items[1].ID)
		assert.Equal(t, cc.ID, items[2].ID)
	})

	t.Run("excludes clarified and discarded items", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		live, err := db.CreateItem(c, gtd.Item{Title: "live"})
		require.NoError(t, err)
		discarded, err := db.CreateItem(c, gtd.Item{Title: "trash"})
		require.NoError(t, err)
		clarified, err := db.CreateItem(c, gtd.Item{Title: "did it"})
		require.NoError(t, err)

		task, err := db.CreateTask(c, gtd.Task{Title: "did it", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		_, err = db.UpdateItemDiscarded(c, discarded.ID)
		require.NoError(t, err)
		_, err = db.UpdateItemClarifiedIntoTask(c, clarified.ID, task.ID)
		require.NoError(t, err)

		items, err := db.ListItems(c)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, live.ID, items[0].ID)
	})

	t.Run("empty inbox", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)
		items, err := db.ListItems(c)
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}

func TestDB_GetItem(t *testing.T) {
	t.Run("returns item with clarify pointer", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		item, err := db.CreateItem(c, gtd.Item{Title: "do something"})
		require.NoError(t, err)
		task, err := db.CreateTask(c, gtd.Task{Title: "do something", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)
		_, err = db.UpdateItemClarifiedIntoTask(c, item.ID, task.ID)
		require.NoError(t, err)

		got, err := db.GetItem(c, item.ID)
		require.NoError(t, err)
		require.NotNil(t, got.ClarifiedIntoTaskID)
		assert.Equal(t, task.ID, *got.ClarifiedIntoTaskID)
		assert.Nil(t, got.ClarifiedIntoProjectID)
		assert.False(t, got.Discarded)
	})

	t.Run("missing item returns error", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)
		_, err := db.GetItem(c, 99999)
		require.Error(t, err)
	})
}

func TestDB_ItemMutualExclusionCheck(t *testing.T) {
	t.Run("rejects clarified-into-task plus discarded", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		item, err := db.CreateItem(c, gtd.Item{Title: "x"})
		require.NoError(t, err)
		task, err := db.CreateTask(c, gtd.Task{Title: "x", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)
		_, err = db.UpdateItemClarifiedIntoTask(c, item.ID, task.ID)
		require.NoError(t, err)

		// Now try to also discard — CHECK should reject.
		_, err = db.UpdateItemDiscarded(c, item.ID)
		require.Error(t, err)
	})

	t.Run("rejects task+project simultaneously", func(t *testing.T) {
		db := openTestDB(t)
		c := ctx(t)

		item, err := db.CreateItem(c, gtd.Item{Title: "x"})
		require.NoError(t, err)
		task, err := db.CreateTask(c, gtd.Task{Title: "x", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)
		proj, err := db.CreateProject(c, gtd.Project{Title: "x", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)

		_, err = db.UpdateItemClarifiedIntoTask(c, item.ID, task.ID)
		require.NoError(t, err)
		_, err = db.UpdateItemClarifiedIntoProject(c, item.ID, proj.ID)
		require.Error(t, err)
	})
}
