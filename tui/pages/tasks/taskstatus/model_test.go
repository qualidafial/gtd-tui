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

// recordingService records the instant passed to ReopenTask; other methods
// are unused (nil embedded interface) and must not be reached by the test
// path.
type recordingService struct {
	gtd.TaskService
	gotID int64
	gotAt time.Time
}

func (s *recordingService) ReopenTask(_ context.Context, id int64, at time.Time) (gtd.Task, error) {
	s.gotID = id
	s.gotAt = at
	return gtd.Task{ID: id, Status: gtd.TaskStatusOpen, StatusChangedAt: at}, nil
}

func typeString(s screen.Screen, text string) screen.Screen {
	for _, r := range text {
		s, _ = s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return s
}

// TestReopen_EditedTimestampReachesService verifies the edited When value
// makes it into the service call. The When field is focused first and
// prefilled with now; the test clears it, types an explicit instant, and
// inspects the model's applyCmd to confirm the value flowed through the
// form to the service.
func TestReopen_EditedTimestampReachesService(t *testing.T) {
	svc := &recordingService{}
	var s screen.Screen = New(gtd.Task{ID: 7, Title: "Resurrect", Status: gtd.TaskStatusDropped}, svc, Reopen)
	s.Init()

	// Clear the prefilled When and type an explicit earlier instant today.
	for range 25 {
		s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	s = typeString(s, "2026-05-24 15:50")

	m, ok := s.(Model)
	require.True(t, ok)
	want := time.Date(2026, 5, 24, 15, 50, 0, 0, time.Local)

	// applyCmd reads the When value from the form and calls the service
	// with that instant.
	msg := m.applyCmd()()
	require.IsType(t, taskTransitionedMsg{}, msg)
	require.NoError(t, msg.(taskTransitionedMsg).err)

	assert.Equal(t, int64(7), svc.gotID)
	assert.WithinDuration(t, want, svc.gotAt, time.Second,
		"service should receive the edited When instant, not the overlay-open time")
}
