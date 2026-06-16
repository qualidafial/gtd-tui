package projectlist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qualidafial/gtd-tui"
)

var testNow = time.Date(2026, 5, 27, 12, 0, 0, 0, time.Local)

func TestProjectStatusMarker(t *testing.T) {
	cases := []struct {
		status gtd.ProjectStatus
		want   string
	}{
		{gtd.ProjectStatusOpen, "[ ]"},
		{gtd.ProjectStatusSomeday, "[?]"},
		{gtd.ProjectStatusDone, "[x]"},
		{gtd.ProjectStatusDropped, "[-]"},
	}
	for _, tc := range cases {
		t.Run(string(tc.status), func(t *testing.T) {
			assert.Equal(t, tc.want, projectStatusMarker(tc.status))
		})
	}
}

func TestProjectChips_TaskProgress(t *testing.T) {
	colors := newChipColors(true)

	t.Run("shows complete/total chip when total > 0", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen}
		counts := gtd.ProjectTaskCounts{Complete: 3, Total: 5}
		chips := projectChips(p, counts, testNow, colors)
		assert.Contains(t, chipTexts(chips), "3/5 tasks")
	})

	t.Run("hides chip when total is 0", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen}
		chips := projectChips(p, gtd.ProjectTaskCounts{}, testNow, colors)
		for _, ch := range chips {
			assert.NotContains(t, ch.text, "tasks")
		}
	})

	t.Run("nothing started shows 0/N tasks", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen}
		counts := gtd.ProjectTaskCounts{Complete: 0, Total: 4}
		chips := projectChips(p, counts, testNow, colors)
		assert.Contains(t, chipTexts(chips), "0/4 tasks")
	})

	t.Run("all done shows N/N tasks", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen}
		counts := gtd.ProjectTaskCounts{Complete: 4, Total: 4}
		chips := projectChips(p, counts, testNow, colors)
		assert.Contains(t, chipTexts(chips), "4/4 tasks")
	})
}

func TestProjectChips_NeedsActionWarning(t *testing.T) {
	colors := newChipColors(true)

	t.Run("open with no tasks shows warning", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen}
		chips := projectChips(p, gtd.ProjectTaskCounts{Complete: 0, Total: 0}, testNow, colors)
		assert.Contains(t, chipTexts(chips), "needs action")
	})

	t.Run("open with all tasks done shows warning", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen}
		chips := projectChips(p, gtd.ProjectTaskCounts{Complete: 3, Total: 3}, testNow, colors)
		assert.Contains(t, chipTexts(chips), "needs action")
	})

	t.Run("open with pending tasks suppresses warning", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen}
		chips := projectChips(p, gtd.ProjectTaskCounts{Complete: 1, Total: 3}, testNow, colors)
		assert.NotContains(t, chipTexts(chips), "needs action")
	})

	t.Run("someday suppresses warning", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusSomeday}
		chips := projectChips(p, gtd.ProjectTaskCounts{}, testNow, colors)
		assert.NotContains(t, chipTexts(chips), "needs action")
	})

	t.Run("done suppresses warning", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusDone}
		chips := projectChips(p, gtd.ProjectTaskCounts{}, testNow, colors)
		assert.NotContains(t, chipTexts(chips), "needs action")
	})

	t.Run("dropped suppresses warning", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusDropped}
		chips := projectChips(p, gtd.ProjectTaskCounts{}, testNow, colors)
		assert.NotContains(t, chipTexts(chips), "needs action")
	})
}

func TestProjectChips_DueChip(t *testing.T) {
	colors := newChipColors(true)

	t.Run("open with future due shows chip", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen, Due: new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))}
		chips := projectChips(p, gtd.ProjectTaskCounts{Complete: 0, Total: 1}, testNow, colors)
		found := false
		for _, ch := range chips {
			if len(ch.text) >= 4 && ch.text[:4] == "due:" {
				found = true
			}
		}
		assert.True(t, found, "expected due chip for open project with due date")
	})

	t.Run("someday suppresses due chip", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusSomeday, Due: new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))}
		chips := projectChips(p, gtd.ProjectTaskCounts{}, testNow, colors)
		for _, ch := range chips {
			assert.NotContains(t, ch.text, "due:", "someday project should not show due chip")
			assert.NotContains(t, ch.text, "overdue:")
		}
	})

	t.Run("done suppresses due chip", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusDone, Due: new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))}
		chips := projectChips(p, gtd.ProjectTaskCounts{}, testNow, colors)
		for _, ch := range chips {
			assert.NotContains(t, ch.text, "due:")
			assert.NotContains(t, ch.text, "overdue:")
		}
	})

	t.Run("open overdue shows overdue chip", func(t *testing.T) {
		p := gtd.Project{Status: gtd.ProjectStatusOpen, Due: new(time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local))}
		chips := projectChips(p, gtd.ProjectTaskCounts{Complete: 0, Total: 1}, testNow, colors)
		found := false
		for _, ch := range chips {
			if len(ch.text) >= 8 && ch.text[:8] == "overdue:" {
				found = true
			}
		}
		assert.True(t, found, "expected overdue chip for past-due open project")
	})
}

func chipTexts(chips []chip) []string {
	out := make([]string, len(chips))
	for i, ch := range chips {
		out[i] = ch.text
	}
	return out
}
