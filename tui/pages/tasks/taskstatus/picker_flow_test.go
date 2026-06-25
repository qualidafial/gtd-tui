package taskstatus

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

// recordingTaskService records which transition the picker invoked.
type recordingTaskService struct {
	gtd.TaskService
	completed, dropped, reopened bool
	gotID                        int64
}

func (s *recordingTaskService) CompleteTask(_ context.Context, id int64, at time.Time) (gtd.Task, error) {
	s.completed, s.gotID = true, id
	return gtd.Task{ID: id, Status: gtd.TaskStatusDone, StatusChangedAt: at}, nil
}

func (s *recordingTaskService) DropTask(_ context.Context, id int64, at time.Time) (gtd.Task, error) {
	s.dropped, s.gotID = true, id
	return gtd.Task{ID: id, Status: gtd.TaskStatusDropped, StatusChangedAt: at}, nil
}

func (s *recordingTaskService) ReopenTask(_ context.Context, id int64, at time.Time) (gtd.Task, error) {
	s.reopened, s.gotID = true, id
	return gtd.Task{ID: id, Status: gtd.TaskStatusOpen, StatusChangedAt: at}, nil
}

func (s *recordingTaskService) any() bool { return s.completed || s.dropped || s.reopened }

func TestPicker_NoOpOnUnchanged(t *testing.T) {
	svc := &recordingTaskService{}
	var s screen.Screen = NewPicker(gtd.Task{ID: 9, Title: "T", Status: gtd.TaskStatusOpen}, svc)
	s.Init()

	// Confirm without moving off the preselected current status.
	_, cmd := s.Update(form.SubmittedMsg{})
	require.NotNil(t, cmd, "submit should emit a dismiss cmd")
	_, ok := cmd().(screen.DismissMsg)
	assert.True(t, ok, "confirming the unchanged status should dismiss")
	assert.False(t, svc.any(), "no transition should run for the unchanged status")
}

func TestPicker_AppliesSelectedTransition(t *testing.T) {
	svc := &recordingTaskService{}
	var s screen.Screen = NewPicker(gtd.Task{ID: 9, Title: "T", Status: gtd.TaskStatusOpen}, svc)
	s.Init()

	// Arrow off "open" onto "done" (the first reachable target), then confirm.
	s, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd := s.Update(form.SubmittedMsg{})
	require.NotNil(t, cmd, "a changed selection should apply a transition")

	msg := cmd()
	require.IsType(t, taskTransitionedMsg{}, msg)
	require.NoError(t, msg.(taskTransitionedMsg).err)
	assert.True(t, svc.completed, "open→done should call CompleteTask")
	assert.Equal(t, int64(9), svc.gotID)
}

func TestPicker_WhenAppearsOnlyAfterChange(t *testing.T) {
	m := NewPicker(gtd.Task{ID: 1, Title: "T", Status: gtd.TaskStatusOpen}, &recordingTaskService{})

	_, shown := m.form.FieldValues()[fieldWhen]
	assert.False(t, shown, "When should be hidden while the status is unchanged")

	moved, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, shown = moved.(Model).form.FieldValues()[fieldWhen]
	assert.True(t, shown, "When should appear once a different status is chosen")
}
