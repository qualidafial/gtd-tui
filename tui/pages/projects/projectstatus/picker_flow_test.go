package projectstatus

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

// recordingProjectService records which transition the picker invoked.
type recordingProjectService struct {
	gtd.ProjectService
	completed, dropped, parked, reopened bool
	gotID                                int64
}

func (s *recordingProjectService) CompleteProject(_ context.Context, id int64, _ bool, at time.Time) (gtd.Project, error) {
	s.completed, s.gotID = true, id
	return gtd.Project{ID: id, Status: gtd.ProjectStatusDone, StatusChangedAt: at}, nil
}

func (s *recordingProjectService) DropProject(_ context.Context, id int64, _ bool, at time.Time) (gtd.Project, error) {
	s.dropped, s.gotID = true, id
	return gtd.Project{ID: id, Status: gtd.ProjectStatusDropped, StatusChangedAt: at}, nil
}

func (s *recordingProjectService) ParkProject(_ context.Context, id int64, at time.Time) (gtd.Project, error) {
	s.parked, s.gotID = true, id
	return gtd.Project{ID: id, Status: gtd.ProjectStatusSomeday, StatusChangedAt: at}, nil
}

func (s *recordingProjectService) ReopenProject(_ context.Context, id int64, at time.Time) (gtd.Project, error) {
	s.reopened, s.gotID = true, id
	return gtd.Project{ID: id, Status: gtd.ProjectStatusOpen, StatusChangedAt: at}, nil
}

func (s *recordingProjectService) any() bool { return s.completed || s.dropped || s.parked || s.reopened }

func TestPicker_NoOpOnUnchanged(t *testing.T) {
	svc := &recordingProjectService{}
	var s screen.Screen = NewPicker(gtd.Project{ID: 3, Title: "P", Status: gtd.ProjectStatusOpen}, 0, svc)
	s.Init()

	_, cmd := s.Update(form.SubmittedMsg{})
	require.NotNil(t, cmd)
	_, ok := cmd().(screen.DismissMsg)
	assert.True(t, ok, "confirming the unchanged status should dismiss")
	assert.False(t, svc.any(), "no transition should run for the unchanged status")
}

func TestPicker_ParksOnSomeday(t *testing.T) {
	svc := &recordingProjectService{}
	var s screen.Screen = NewPicker(gtd.Project{ID: 3, Title: "P", Status: gtd.ProjectStatusOpen}, 0, svc)
	s.Init()

	// open → someday is the first reachable target.
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd := s.Update(form.SubmittedMsg{})
	require.NotNil(t, cmd)

	msg := cmd()
	require.IsType(t, projectTransitionedMsg{}, msg)
	require.NoError(t, msg.(projectTransitionedMsg).err)
	assert.True(t, svc.parked, "open→someday should call ParkProject")
	assert.Equal(t, int64(3), svc.gotID)
}

func TestPicker_WhenAppearsOnlyAfterChange(t *testing.T) {
	m := NewPicker(gtd.Project{ID: 1, Title: "P", Status: gtd.ProjectStatusOpen}, 0, &recordingProjectService{})

	_, shown := m.form.FieldValues()[fieldWhen]
	assert.False(t, shown, "When should be hidden while the status is unchanged")

	moved, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, shown = moved.(Model).form.FieldValues()[fieldWhen]
	assert.True(t, shown, "When should appear once a different status is chosen")
}
