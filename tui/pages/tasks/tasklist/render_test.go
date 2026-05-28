package tasklist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qualidafial/gtd-tui"
)

// now is a stable reference: 2026-05-24 (Sunday) at 12:00 local.
var now = time.Date(2026, 5, 24, 12, 0, 0, 0, time.Local)

func dateOnly(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func TestTaskChips(t *testing.T) {
	colors := newChipColors(true)

	t.Run("date-only due today is not overdue", func(t *testing.T) {
		// now at 17:00; a date-only due today applies at end of day.
		at := time.Date(2026, 5, 24, 17, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusOpen, Due: new(dateOnly(2026, 5, 24))}
		ch, ok := dueChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "due:today", ch.text)
	})

	t.Run("date-only due flips to overdue next day", func(t *testing.T) {
		at := time.Date(2026, 5, 28, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusOpen, Due: new(dateOnly(2026, 5, 27))}
		ch, ok := dueChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "overdue:1d", ch.text)
	})

	t.Run("timed due earlier today is overdue", func(t *testing.T) {
		at := time.Date(2026, 5, 24, 17, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusOpen, Due: new(time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local))}
		ch, ok := dueChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "overdue:3pm", ch.text)
	})

	t.Run("defer flips to ready at start of day", func(t *testing.T) {
		at := time.Date(2026, 5, 27, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusOpen, DeferUntil: new(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "ready:today", ch.text)
	})

	t.Run("future defer", func(t *testing.T) {
		at := time.Date(2026, 5, 26, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusOpen, DeferUntil: new(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "defer:tomorrow", ch.text)
	})

	t.Run("resurfaced defer counts days since", func(t *testing.T) {
		at := time.Date(2026, 5, 28, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusOpen, DeferUntil: new(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "ready:1d", ch.text)
	})

	t.Run("done hides date chips but keeps assignee", func(t *testing.T) {
		task := gtd.Task{
			Status:   gtd.TaskStatusDone,
			Due:      new(dateOnly(2026, 5, 20)),
			Assignee: new("alice"),
		}
		chips := taskChips(task, now, colors)
		assert.Equal(t, []chip{{text: "@alice", style: colors.assignee}}, chips)
	})

	t.Run("dropped hides all chips", func(t *testing.T) {
		task := gtd.Task{
			Status:   gtd.TaskStatusDropped,
			Due:      new(dateOnly(2026, 5, 20)),
			Assignee: new("bob"),
		}
		assert.Empty(t, taskChips(task, now, colors))
	})

	t.Run("chip order: due, defer, assignee", func(t *testing.T) {
		task := gtd.Task{
			Status:     gtd.TaskStatusOpen,
			Due:        new(dateOnly(2026, 6, 21)), // 28d out
			DeferUntil: new(dateOnly(2026, 6, 7)),  // 14d out
			Assignee:   new("carol"),
		}
		chips := taskChips(task, now, colors)
		var texts []string
		for _, ch := range chips {
			texts = append(texts, ch.text)
		}
		assert.Equal(t, []string{"due:28d", "defer:14d", "@carol"}, texts)
	})

	t.Run("ready day-count and color", func(t *testing.T) {
		at := time.Date(2026, 5, 30, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusOpen, DeferUntil: new(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "ready:3d", ch.text)
		assert.Equal(t, colors.ready, ch.style)
	})
}
