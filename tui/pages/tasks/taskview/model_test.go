package taskview

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskedit"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

// stubScreen is a no-op Screen returned by test factories.
type stubScreen struct{ name string }

func (stubScreen) Init() tea.Cmd                             { return nil }
func (s stubScreen) Update(tea.Msg) (screen.Screen, tea.Cmd) { return s, nil }
func (stubScreen) View() string                              { return "" }
func (stubScreen) Keys() []keymap.Group                      { return nil }

func ptr[T any](v T) *T { return &v }

func newView(task gtd.Task) Model {
	return New(
		task,
		nil,
		func(int64) string { return "Renovate" },
		func(gtd.Task) screen.Screen { return stubScreen{name: "picker"} },
		func(gtd.Task) screen.Screen { return stubScreen{name: "convert"} },
		func(gtd.Project) screen.Screen { return stubScreen{name: "projectview"} },
	)
}

func pushedScreen(t *testing.T, cmd tea.Cmd) screen.Screen {
	t.Helper()
	require.NotNil(t, cmd, "expected a cmd")
	push, ok := cmd().(screen.PushMsg)
	require.True(t, ok, "expected PushMsg, got %T", cmd())
	return push.Screen
}

func TestHeader_AllFields(t *testing.T) {
	due := time.Date(2026, 6, 15, 0, 0, 0, 0, time.Local)
	m := newView(gtd.Task{
		ID:          1,
		Title:       "Patch the roof",
		Status:      gtd.TaskStatusOpen,
		ProjectID:   ptr(int64(7)),
		Assignee:    ptr("bob"),
		Due:         &due,
		Description: "before the rains",
	})
	header := m.renderHeader()

	assert.Contains(t, header, "Patch the roof")
	assert.Contains(t, header, "Open")
	assert.Contains(t, header, "+Renovate")
	assert.Contains(t, header, "bob")
	assert.Contains(t, header, "2026-06-15")
	assert.Contains(t, header, "before the rains")
}

func TestHeader_StandaloneOmitsEmpty(t *testing.T) {
	m := newView(gtd.Task{ID: 1, Title: "Lone task", Status: gtd.TaskStatusOpen})
	header := m.renderHeader()

	assert.Contains(t, header, "Lone task")
	assert.Contains(t, header, "Open")
	assert.NotContains(t, header, "Project:")
	assert.NotContains(t, header, "Assignee:")
	assert.NotContains(t, header, "Due:")
	assert.NotContains(t, header, "Description:")
}

func TestHeader_StatusLabels(t *testing.T) {
	for _, tt := range []struct {
		status gtd.TaskStatus
		label  string
	}{
		{gtd.TaskStatusOpen, "Open"},
		{gtd.TaskStatusDone, "Done"},
		{gtd.TaskStatusDropped, "Dropped"},
	} {
		t.Run(string(tt.status), func(t *testing.T) {
			m := newView(gtd.Task{ID: 1, Title: "T", Status: tt.status})
			assert.Contains(t, m.renderHeader(), tt.label)
		})
	}
}

func TestKeybindings_Guards(t *testing.T) {
	open := newView(gtd.Task{ID: 1, Status: gtd.TaskStatusOpen, ProjectID: ptr(int64(7))})
	assert.True(t, open.KeyMap.Drop.Enabled(), "drop enabled for open task")
	assert.False(t, open.KeyMap.ConvertToProject.Enabled(), "convert disabled for task in a project")
	assert.True(t, open.KeyMap.GoToProject.Enabled(), "go-to-project enabled when task has a project")

	standaloneDone := newView(gtd.Task{ID: 2, Status: gtd.TaskStatusDone})
	assert.False(t, standaloneDone.KeyMap.Drop.Enabled(), "drop disabled for closed task")
	assert.True(t, standaloneDone.KeyMap.ConvertToProject.Enabled(), "convert enabled for standalone task")
	assert.False(t, standaloneDone.KeyMap.GoToProject.Enabled(), "go-to-project disabled for standalone task")
}

func TestToggleLabel_TracksStatus(t *testing.T) {
	open := newView(gtd.Task{ID: 1, Status: gtd.TaskStatusOpen})
	assert.Equal(t, "complete", open.KeyMap.ToggleComplete.Help().Desc)

	done := newView(gtd.Task{ID: 1, Status: gtd.TaskStatusDone})
	assert.Equal(t, "reopen", done.KeyMap.ToggleComplete.Help().Desc)
}

func TestEdit_PushesEditor(t *testing.T) {
	m := newView(gtd.Task{ID: 1, Title: "T", Status: gtd.TaskStatusOpen})
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	_, ok := pushedScreen(t, cmd).(taskedit.Model)
	assert.True(t, ok, "e should push the task editor")
}

func TestToggle_ResolvesTransition(t *testing.T) {
	for _, tt := range []struct {
		status gtd.TaskStatus
		want   taskstatus.Transition
	}{
		{gtd.TaskStatusOpen, taskstatus.Complete},
		{gtd.TaskStatusDone, taskstatus.Reopen},
		{gtd.TaskStatusDropped, taskstatus.Reopen},
	} {
		t.Run(string(tt.status), func(t *testing.T) {
			m := newView(gtd.Task{ID: 1, Status: tt.status})
			_, cmd := m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
			ov, ok := pushedScreen(t, cmd).(taskstatus.Model)
			require.True(t, ok, "space should push a status overlay")
			assert.Equal(t, tt.want, ov.Transition())
		})
	}
}

func TestDrop_PushesDropOverlay(t *testing.T) {
	m := newView(gtd.Task{ID: 1, Status: gtd.TaskStatusOpen})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	ov, ok := pushedScreen(t, cmd).(taskstatus.Model)
	require.True(t, ok, "delete should push a status overlay")
	assert.Equal(t, taskstatus.Drop, ov.Transition())
}

func TestDrop_InertOnClosedTask(t *testing.T) {
	m := newView(gtd.Task{ID: 1, Status: gtd.TaskStatusDone})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	assert.Nil(t, cmd, "delete on a closed task should be inert")
}

func TestAssignAndConvert_PushFactories(t *testing.T) {
	standalone := newView(gtd.Task{ID: 1, Status: gtd.TaskStatusOpen})

	_, cmd := standalone.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	if s, ok := pushedScreen(t, cmd).(stubScreen); assert.True(t, ok) {
		assert.Equal(t, "picker", s.name)
	}

	_, cmd = standalone.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if s, ok := pushedScreen(t, cmd).(stubScreen); assert.True(t, ok) {
		assert.Equal(t, "convert", s.name)
	}
}

func TestGoToProject_ReplacesWithProjectView(t *testing.T) {
	m := newView(gtd.Task{ID: 1, Status: gtd.TaskStatusOpen, ProjectID: ptr(int64(7))})
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})

	s, ok := next.(stubScreen)
	require.True(t, ok, "g should replace the view with the project view; got %T", next)
	assert.Equal(t, "projectview", s.name)
	assert.NotNil(t, cmd, "Replace batches the next screen's init")
}

func TestReload_UpdatesTask(t *testing.T) {
	m := newView(gtd.Task{ID: 1, Title: "Old", Status: gtd.TaskStatusOpen})
	next, _ := m.Update(taskReloadedMsg{task: gtd.Task{ID: 1, Title: "New", Status: gtd.TaskStatusDone}})
	assert.Contains(t, next.(Model).renderHeader(), "New")
	assert.Contains(t, next.(Model).renderHeader(), "Done")
}
