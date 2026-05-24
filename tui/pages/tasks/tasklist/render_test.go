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

func TestFormatWhen(t *testing.T) {
	tests := []struct {
		name string
		ref  time.Time
		want string
	}{
		{"date-only today", dateOnly(2026, 5, 24), "today"},
		{"tomorrow", dateOnly(2026, 5, 25), "tomorrow"},
		{"weekday band low (2d)", dateOnly(2026, 5, 26), "tuesday"},
		{"weekday band high (6d)", dateOnly(2026, 5, 30), "saturday"},
		{"seven days", dateOnly(2026, 5, 31), "7d"},
		{"thirty days", dateOnly(2026, 6, 23), "30d"},
		{"thirty-one days absolute", dateOnly(2026, 6, 24), "2026-06-24"},
		{"timed today still ahead", time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local), "3pm"},
		{"timed earlier today", time.Date(2026, 5, 24, 9, 30, 0, 0, time.Local), "9:30am"},
		{"past three days", dateOnly(2026, 5, 21), "3d"},
		{"past thirty days", dateOnly(2026, 4, 24), "30d"},
		{"past thirty-one days absolute", dateOnly(2026, 4, 23), "2026-04-23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatWhen(tt.ref, now))
		})
	}
}

func TestTaskChips(t *testing.T) {
	colors := newChipColors(true)
	ptr := func(tm time.Time) *time.Time { return &tm }

	t.Run("date-only due today is not overdue", func(t *testing.T) {
		// now at 17:00; a date-only due today applies at end of day.
		at := time.Date(2026, 5, 24, 17, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusPending, Due: ptr(dateOnly(2026, 5, 24))}
		ch, ok := dueChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "due:today", ch.text)
	})

	t.Run("date-only due flips to overdue next day", func(t *testing.T) {
		at := time.Date(2026, 5, 28, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusPending, Due: ptr(dateOnly(2026, 5, 27))}
		ch, ok := dueChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "overdue:1d", ch.text)
	})

	t.Run("timed due earlier today is overdue", func(t *testing.T) {
		at := time.Date(2026, 5, 24, 17, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusPending, Due: ptr(time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local))}
		ch, ok := dueChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "overdue:3pm", ch.text)
	})

	t.Run("defer flips to ready at start of day", func(t *testing.T) {
		at := time.Date(2026, 5, 27, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusPending, DeferUntil: ptr(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "ready:today", ch.text)
	})

	t.Run("future defer", func(t *testing.T) {
		at := time.Date(2026, 5, 26, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusPending, DeferUntil: ptr(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "defer:tomorrow", ch.text)
	})

	t.Run("resurfaced defer counts days since", func(t *testing.T) {
		at := time.Date(2026, 5, 28, 9, 0, 0, 0, time.Local)
		task := gtd.Task{Status: gtd.TaskStatusPending, DeferUntil: ptr(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "ready:1d", ch.text)
	})

	t.Run("done hides date chips but keeps assignee", func(t *testing.T) {
		task := gtd.Task{
			Status:   gtd.TaskStatusDone,
			Due:      ptr(dateOnly(2026, 5, 20)),
			Assignee: "alice",
		}
		chips := taskChips(task, now, colors)
		assert.Equal(t, []chip{{text: "@alice", style: colors.assignee}}, chips)
	})

	t.Run("dropped hides all chips", func(t *testing.T) {
		task := gtd.Task{
			Status:   gtd.TaskStatusDropped,
			Due:      ptr(dateOnly(2026, 5, 20)),
			Assignee: "bob",
		}
		assert.Empty(t, taskChips(task, now, colors))
	})

	t.Run("chip order: due, defer, assignee", func(t *testing.T) {
		task := gtd.Task{
			Status:     gtd.TaskStatusPending,
			Due:        ptr(dateOnly(2026, 6, 21)), // 28d out
			DeferUntil: ptr(dateOnly(2026, 6, 7)),  // 14d out
			Assignee:   "carol",
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
		task := gtd.Task{Status: gtd.TaskStatusPending, DeferUntil: ptr(dateOnly(2026, 5, 27))}
		ch, ok := deferChip(task, at, colors)
		assert.True(t, ok)
		assert.Equal(t, "ready:3d", ch.text)
		assert.Equal(t, colors.ready, ch.style)
	})
}
