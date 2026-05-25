package taskstatus

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

// recordingService records the instant passed to ReopenTask; other methods are
// unused (nil embedded interface) and must not be reached by the test path.
type recordingService struct {
	gtd.TaskService
	gotID int64
	gotAt time.Time
}

func (s *recordingService) ReopenTask(_ context.Context, id int64, at time.Time) (gtd.Task, error) {
	s.gotID = id
	s.gotAt = at
	return gtd.Task{ID: id, Status: gtd.TaskStatusPending, StatusChangedAt: at}, nil
}

func typeString(s screen.Screen, text string) screen.Screen {
	for _, r := range text {
		s, _ = s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return s
}

// TestReopen_EditedTimestampReachesService verifies the edited When value
// survives the Model copies that flow through the screen stack and is the
// instant handed to the service — the stale-copy binding regression. The date
// field is bound to a shared **time.Time slot, so a copy distinct from the one
// New() built must still observe the edit and applyCmd must read it.
func TestReopen_EditedTimestampReachesService(t *testing.T) {
	svc := &recordingService{}
	var s screen.Screen = New(gtd.Task{ID: 7, Title: "Resurrect", Status: gtd.TaskStatusDropped}, svc, Reopen)
	s.Init()

	// The When date field is focused first and prefilled with now. Clear it and
	// type an explicit earlier instant today.
	for range 25 {
		s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	s = typeString(s, "2026-05-24 15:50")

	// Enter commits the date field (parse + write to the shared slot). s is now
	// a copy several Updates removed from the model New() constructed.
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m, ok := s.(Model)
	require.True(t, ok)
	want := time.Date(2026, 5, 24, 15, 50, 0, 0, time.Local)

	// The edited value is visible through the shared slot on this copy.
	require.NotNil(t, *m.at)
	assert.WithinDuration(t, want, **m.at, time.Second,
		"copy should observe the edited When via the shared slot")

	// applyCmd (the method the confirm path invokes) reads that slot and calls
	// the service with the edited instant, not the overlay-open time.
	msg := m.applyCmd()()
	require.IsType(t, taskTransitionedMsg{}, msg)
	require.NoError(t, msg.(taskTransitionedMsg).err)

	assert.Equal(t, int64(7), svc.gotID)
	assert.WithinDuration(t, want, svc.gotAt, time.Second,
		"service should receive the edited When instant, not the overlay-open time")
}
