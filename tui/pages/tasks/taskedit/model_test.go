package taskedit

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
)

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestModel_StatusLine(t *testing.T) {
	tests := []struct {
		name   string
		status gtd.TaskStatus
		at     time.Time
		want   string
	}{
		{
			name:   "open changed three days ago",
			status: gtd.TaskStatusOpen,
			at:     time.Now().AddDate(0, 0, -3),
			want:   "Status:  Open (3d)",
		},
		{
			name:   "done changed today",
			status: gtd.TaskStatusDone,
			at:     time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
			want:   "Status:  Done (today)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(gtd.Task{ID: 1, Title: "Existing", Status: tt.status, StatusChangedAt: tt.at}, nil, "")
			view := ansi.Strip(m.View())
			if !strings.Contains(view, tt.want) {
				t.Fatalf("expected status line %q in view, got:\n%s", tt.want, view)
			}
		})
	}
}

func TestModel_SaveError_ReturnsErrorCmd(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "")

	_, cmd := m.Update(taskSavedMsg{err: errors.New("disk full")})
	require.NotNil(t, cmd, "expected error cmd on save failure")
	msg := cmd()
	err, ok := msg.(error)
	require.True(t, ok, "expected error msg, got %T", msg)
	assert.Contains(t, err.Error(), "disk full")
}

func TestModel_SaveError_EscClearsError(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "")

	withErr, _ := m.Update(taskSavedMsg{err: errors.New("disk full")})
	require.NotNil(t, withErr.(Model).err, "precondition: error must be set")

	cleared, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Nil(t, cmd)
	assert.Nil(t, cleared.(Model).err)
}

func TestModel_SaveError_OtherKeysSwallowed(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "")

	withErr, _ := m.Update(taskSavedMsg{err: errors.New("disk full")})

	_, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.Nil(t, cmd)
}

func TestModel_ProjectLine_Shown(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "Inbox Rewrite")
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Project: Inbox Rewrite") {
		t.Fatalf("expected project line in view, got:\n%s", view)
	}
}

func TestModel_ProjectLine_Hidden(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "")
	view := ansi.Strip(m.View())
	if strings.Contains(view, "Project:") {
		t.Fatalf("expected no project line in view, got:\n%s", view)
	}
}

func TestCtrlEnter_SavesFromTitleField(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewTaskService(db)
	created, err := svc.CreateTask(t.Context(), gtd.Task{Title: "Original", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	var s screen.Screen = New(created, svc, "")
	s = screentest.Init(t, s)

	// Edit the title from the Title field, then submit via ctrl+s without
	// tabbing through Description / Assignee / Due / Defer.
	s = screentest.TypeText(t, s, " edited")

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.True(t, dismissed, "expected overlay to dismiss after ctrl+s save")

	got, err := svc.GetTask(t.Context(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Original edited", got.Title)
}

func TestCtrlEnter_EmptyTitle_NoSave(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewTaskService(db)

	var s screen.Screen = New(gtd.Task{}, svc, "")
	s = screentest.Init(t, s)

	// Title is empty; ctrl+s must not dismiss or create a task.
	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.False(t, dismissed, "overlay must not dismiss with empty title")

	tasks, err := svc.ListTasks(t.Context(), gtd.TaskFilter{})
	require.NoError(t, err)
	assert.Empty(t, tasks, "no task should be created")
}
