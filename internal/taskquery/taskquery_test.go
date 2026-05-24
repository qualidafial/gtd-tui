package taskquery

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
)

// fixedNow is a stable reference instant for date-resolution tests.
var fixedNow = time.Date(2026, 5, 24, 9, 30, 0, 0, time.Local)

func endOfDay(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 23, 59, 59, int(time.Second-time.Nanosecond), time.Local).UTC()
}

func startOfDay(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).UTC()
}

func TestParse_StructuredKeys(t *testing.T) {
	pending := gtd.TaskStatusPending
	done := gtd.TaskStatusDone
	dropped := gtd.TaskStatusDropped
	nextAction := gtd.TaskKindNextAction
	delegated := gtd.TaskKindDelegated
	bob := "bob"

	tests := []struct {
		name  string
		query string
		want  gtd.TaskFilter
	}{
		{"status pending", "status:pending", gtd.TaskFilter{Status: &pending}},
		{"status done", "status:done", gtd.TaskFilter{Status: &done}},
		{"status dropped", "status:dropped", gtd.TaskFilter{Status: &dropped}},
		{"kind next_action", "kind:next_action", gtd.TaskFilter{Kind: &nextAction}},
		{"kind delegated", "kind:delegated", gtd.TaskFilter{Kind: &delegated}},
		{"assignee", "assignee:bob", gtd.TaskFilter{Assignee: &bob}},
		{
			"ready now is instant",
			"ready:now",
			gtd.TaskFilter{Ready: &gtd.DatePredicate{Kind: gtd.AvailableAsOf, Time: fixedNow.UTC()}},
		},
		{
			"due threshold",
			"due:today",
			gtd.TaskFilter{Due: &gtd.DatePredicate{Kind: gtd.OnOrBefore, Time: endOfDay(2026, 5, 24)}},
		},
		{
			"defer threshold",
			"defer:0d",
			gtd.TaskFilter{Defer: &gtd.DatePredicate{Kind: gtd.After, Time: startOfDay(2026, 5, 24)}},
		},
		{
			"due none",
			"due:none",
			gtd.TaskFilter{Due: &gtd.DatePredicate{Kind: gtd.IsNull}},
		},
		{
			"defer any",
			"defer:any",
			gtd.TaskFilter{Defer: &gtd.DatePredicate{Kind: gtd.IsNotNull}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAt(tt.query, fixedNow)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParse_FreeText(t *testing.T) {
	t.Run("bare word", func(t *testing.T) {
		got, err := parseAt("bob", fixedNow)
		require.NoError(t, err)
		assert.Equal(t, []string{"bob"}, got.Search)
	})

	t.Run("unrecognized key is free text", func(t *testing.T) {
		got, err := parseAt("foo:bar", fixedNow)
		require.NoError(t, err)
		assert.Equal(t, []string{"foo:bar"}, got.Search)
	})

	t.Run("multiple terms ANDed", func(t *testing.T) {
		got, err := parseAt("report bob", fixedNow)
		require.NoError(t, err)
		assert.Equal(t, []string{"report", "bob"}, got.Search)
	})

	t.Run("mixed structured and free text", func(t *testing.T) {
		got, err := parseAt("status:done report", fixedNow)
		require.NoError(t, err)
		require.NotNil(t, got.Status)
		assert.Equal(t, gtd.TaskStatusDone, *got.Status)
		assert.Equal(t, []string{"report"}, got.Search)
	})
}

func TestParse_LastWins(t *testing.T) {
	got, err := parseAt("status:done status:dropped", fixedNow)
	require.NoError(t, err)
	require.NotNil(t, got.Status)
	assert.Equal(t, gtd.TaskStatusDropped, *got.Status)
}

func TestParse_DateValues(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  time.Time
	}{
		// due thresholds at end-of-day.
		{"due relative future days", "due:7d", endOfDay(2026, 5, 31)},
		{"due relative past days", "due:-5d", endOfDay(2026, 5, 19)},
		{"due overdue alias", "due:overdue", endOfDay(2026, 5, 23)},
		{"due today alias", "due:today", endOfDay(2026, 5, 24)},
		{"due week alias", "due:week", endOfDay(2026, 5, 31)},
		{"due iso date", "due:2026-06-01", endOfDay(2026, 6, 1)},
		// defer/ready threshold at start-of-day.
		{"defer week unit", "defer:2w", startOfDay(2026, 6, 7)},
		{"defer today alias", "defer:today", startOfDay(2026, 5, 24)},
		{"defer relative future days", "defer:7d", startOfDay(2026, 5, 31)},
		{"defer iso date", "defer:2026-06-01", startOfDay(2026, 6, 1)},
		{"ready today alias", "ready:today", startOfDay(2026, 5, 24)},
		{"ready relative past days", "ready:-3d", startOfDay(2026, 5, 21)},
		{"ready iso date", "ready:2026-06-01", startOfDay(2026, 6, 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAt(tt.query, fixedNow)
			require.NoError(t, err)
			var pred *gtd.DatePredicate
			switch {
			case got.Due != nil:
				pred = got.Due
			case got.Defer != nil:
				pred = got.Defer
			case got.Ready != nil:
				pred = got.Ready
			}
			require.NotNil(t, pred)
			assert.True(t, pred.Time.Equal(tt.want), "got %v want %v", pred.Time, tt.want)
		})
	}

	t.Run("now is the instant not end of day", func(t *testing.T) {
		got, err := parseAt("ready:now", fixedNow)
		require.NoError(t, err)
		require.NotNil(t, got.Ready)
		assert.True(t, got.Ready.Time.Equal(fixedNow.UTC()))
	})

	t.Run("none and any", func(t *testing.T) {
		got, err := parseAt("defer:none", fixedNow)
		require.NoError(t, err)
		require.NotNil(t, got.Defer)
		assert.Equal(t, gtd.IsNull, got.Defer.Kind)

		got, err = parseAt("defer:any", fixedNow)
		require.NoError(t, err)
		require.NotNil(t, got.Defer)
		assert.Equal(t, gtd.IsNotNull, got.Defer.Kind)
	})
}

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantStart int
		wantEnd   int
	}{
		{"bad status", "status:bogus", 0, 12},
		{"bad kind", "kind:foo", 0, 8},
		{"bad date", "kind:delegated due:notadate", 15, 27},
		{"ready none not allowed", "ready:none", 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAt(tt.query, fixedNow)
			require.Error(t, err)
			var pe *ParseError
			require.True(t, errors.As(err, &pe), "expected *ParseError, got %T", err)
			assert.Equal(t, tt.wantStart, pe.Start, "start offset")
			assert.Equal(t, tt.wantEnd, pe.End, "end offset")
		})
	}
}

func TestParse_Empty(t *testing.T) {
	got, err := parseAt("", fixedNow)
	require.NoError(t, err)
	assert.Equal(t, gtd.TaskFilter{}, got)

	got, err = parseAt("   ", fixedNow)
	require.NoError(t, err)
	assert.Equal(t, gtd.TaskFilter{}, got)
}

func TestParse_TimezoneBoundaries(t *testing.T) {
	// Across a range of fixed local offsets, `now` stays the exact instant
	// while `today` snaps to end-of-local-day.
	for _, offset := range []int{-8, -1, 0, 1, 5, 13} {
		loc := time.FixedZone("test", offset*3600)
		ref := time.Date(2026, 5, 24, 9, 30, 0, 0, loc)

		// now: exact instant regardless of zone.
		f, err := parseAtLoc("ready:now", ref, loc)
		require.NoError(t, err)
		require.NotNil(t, f.Ready)
		assert.True(t, f.Ready.Time.Equal(ref.UTC()), "offset %d: now should be the instant", offset)

		// today: end-of-day in loc.
		f, err = parseAtLoc("due:today", ref, loc)
		require.NoError(t, err)
		require.NotNil(t, f.Due)
		wantEOD := time.Date(2026, 5, 24, 23, 59, 59, int(time.Second-time.Nanosecond), loc).UTC()
		assert.True(t, f.Due.Time.Equal(wantEOD), "offset %d: today should be end-of-local-day", offset)
	}
}
