package taskpicker

import (
	"reflect"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
)

func setup(t *testing.T) (gtd.TaskService, gtd.ProjectService) {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return service.NewTaskService(db), service.NewProjectService(db)
}

// flatten runs cmd and returns every leaf message it produces, expanding
// tea.Batch and tea.Sequence (whose sequenceMsg is an unexported []tea.Cmd, so
// it is walked via reflection).
func flatten(t *testing.T, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	var out []tea.Msg
	if cmd == nil {
		return out
	}
	msg := cmd()
	if msg == nil {
		return out
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			out = append(out, flatten(t, c)...)
		}
		return out
	}
	if v := reflect.ValueOf(msg); v.Kind() == reflect.Slice {
		// Likely a sequenceMsg ([]tea.Cmd); walk it.
		expanded := true
		for i := 0; i < v.Len(); i++ {
			c, ok := v.Index(i).Interface().(tea.Cmd)
			if !ok {
				expanded = false
				break
			}
			out = append(out, flatten(t, c)...)
		}
		if expanded {
			return out
		}
	}
	return append(out, msg)
}

func selectedMsgIn(msgs []tea.Msg) (SelectedMsg, bool) {
	for _, m := range msgs {
		if sm, ok := m.(SelectedMsg); ok {
			return sm, true
		}
	}
	return SelectedMsg{}, false
}

func hasDismiss(msgs []tea.Msg) bool {
	for _, m := range msgs {
		if _, ok := m.(screen.DismissMsg); ok {
			return true
		}
	}
	return false
}

func TestPicker_ListsOnlyStandaloneOpenTasks(t *testing.T) {
	svc, projSvc := setup(t)
	ctx := t.Context()

	standalone, err := svc.CreateTask(ctx, gtd.Task{Title: "loose end", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)
	// A task already in a project must not be offered.
	proj, err := projSvc.CreateProject(ctx, gtd.Project{Title: "host", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)
	_, err = svc.CreateTask(ctx, gtd.Task{Title: "owned task", Status: gtd.TaskStatusOpen, ProjectID: &proj.ID})
	require.NoError(t, err)
	// A closed standalone task must not be offered.
	done, err := svc.CreateTask(ctx, gtd.Task{Title: "finished thing", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)
	_, err = svc.CompleteTask(ctx, done.ID, time.Now())
	require.NoError(t, err)

	var m screen.Screen = New(svc)
	m = screentest.Init(t, m)

	// White-box: the loaded candidate set is exactly the standalone open tasks.
	pm, ok := m.(Model)
	require.True(t, ok)
	require.Len(t, pm.tasks, 1)
	assert.Equal(t, standalone.ID, pm.tasks[0].ID)
	assert.Equal(t, "loose end", pm.tasks[0].Title)
}

func TestPicker_ConfirmEmitsSelectionAndDismisses(t *testing.T) {
	svc, _ := setup(t)
	ctx := t.Context()

	task, err := svc.CreateTask(ctx, gtd.Task{Title: "pick me", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	m := New(svc)
	var s screen.Screen = screentest.Init(t, m)

	// Confirm the (only, pre-selected) option.
	next, cmd := s.Update(form.SubmittedMsg{})
	msgs := flatten(t, cmd)

	sm, ok := selectedMsgIn(msgs)
	require.True(t, ok, "confirm should broadcast a SelectedMsg")
	assert.Equal(t, task.ID, sm.Task.ID)
	assert.Equal(t, "pick me", sm.Task.Title)
	assert.True(t, hasDismiss(msgs), "confirm should also dismiss the overlay")

	// The picker performs no mutation: the task is still standalone.
	got, err := svc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Nil(t, got.ProjectID)
	_ = next
}

func TestPicker_CancelDoesNotEmit(t *testing.T) {
	svc, _ := setup(t)
	ctx := t.Context()

	_, err := svc.CreateTask(ctx, gtd.Task{Title: "pick me", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	var m screen.Screen = New(svc)
	m = screentest.Init(t, m)

	var msgs []tea.Msg
	var dismissed bool
	for s, msg := range screentest.PumpSend(t, m, tea.KeyPressMsg{Code: tea.KeyEscape}) {
		m = s
		if msg != nil {
			msgs = append(msgs, msg)
		}
		if _, ok := msg.(screen.DismissMsg); ok {
			dismissed = true
		}
	}
	require.True(t, dismissed, "esc should dismiss")
	_, ok := selectedMsgIn(msgs)
	assert.False(t, ok, "esc must not broadcast a selection")
}
