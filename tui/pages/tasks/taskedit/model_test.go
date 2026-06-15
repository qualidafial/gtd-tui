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
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
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
			m := New(gtd.Task{ID: 1, Title: "Existing", Status: tt.status, StatusChangedAt: tt.at}, nil, "", nil)
			view := ansi.Strip(m.View())
			if !strings.Contains(view, tt.want) {
				t.Fatalf("expected status line %q in view, got:\n%s", tt.want, view)
			}
		})
	}
}

func TestModel_SaveError_ReturnsErrorCmd(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "", nil)

	_, cmd := m.Update(taskSavedMsg{err: errors.New("disk full")})
	require.NotNil(t, cmd, "expected error cmd on save failure")
	msg := cmd()
	err, ok := msg.(error)
	require.True(t, ok, "expected error msg, got %T", msg)
	assert.Contains(t, err.Error(), "disk full")
}

func TestModel_SaveError_EscClearsError(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "", nil)

	withErr, _ := m.Update(taskSavedMsg{err: errors.New("disk full")})
	require.NotNil(t, withErr.(Model).err, "precondition: error must be set")

	cleared, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Nil(t, cmd)
	assert.Nil(t, cleared.(Model).err)
}

func TestModel_SaveError_OtherKeysSwallowed(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "", nil)

	withErr, _ := m.Update(taskSavedMsg{err: errors.New("disk full")})

	_, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.Nil(t, cmd)
}

// stubScreen is a no-op Screen used as the result of a test view factory.
type stubScreen struct{}

func (stubScreen) Init() tea.Cmd                           { return nil }
func (stubScreen) Update(tea.Msg) (screen.Screen, tea.Cmd) { return stubScreen{}, nil }
func (stubScreen) View() string                            { return "" }
func (stubScreen) Keys() []keymap.Group                    { return nil }

func TestModel_Save_CreateWithFactory_ReplacesWithView(t *testing.T) {
	var gotTask gtd.Task
	factory := func(task gtd.Task) screen.Screen {
		gotTask = task
		return stubScreen{}
	}
	m := New(gtd.Task{}, nil, "", factory)

	created := gtd.Task{ID: 7, Title: "Fresh"}
	next, cmd := m.Update(taskSavedMsg{task: created, created: true})

	assert.NotNil(t, cmd, "expected a cmd batching window-size + init")
	_, ok := next.(stubScreen)
	assert.True(t, ok, "create with factory should replace the editor with the task view; got %T", next)
	assert.Equal(t, created, gotTask, "factory should receive the created task")
}

func TestModel_Save_CreateWithoutFactory_Dismisses(t *testing.T) {
	m := New(gtd.Task{}, nil, "", nil)

	_, cmd := m.Update(taskSavedMsg{task: gtd.Task{ID: 7}, created: true})
	require.NotNil(t, cmd)
	_, ok := cmd().(screen.DismissMsg)
	assert.True(t, ok, "create without factory should dismiss")
}

func TestModel_Save_Update_Dismisses(t *testing.T) {
	factory := func(gtd.Task) screen.Screen { return stubScreen{} }
	m := New(gtd.Task{ID: 7, Title: "Existing"}, nil, "", factory)

	_, cmd := m.Update(taskSavedMsg{task: gtd.Task{ID: 7}, created: false})
	require.NotNil(t, cmd)
	_, ok := cmd().(screen.DismissMsg)
	assert.True(t, ok, "update should dismiss even with a factory")
}

func TestModel_ProjectLine_Shown(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "Inbox Rewrite", nil)
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Project: Inbox Rewrite") {
		t.Fatalf("expected project line in view, got:\n%s", view)
	}
}

func TestModel_ProjectLine_Hidden(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil, "", nil)
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

	var s screen.Screen = New(created, svc, "", nil)
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

	var s screen.Screen = New(gtd.Task{}, svc, "", nil)
	s = screentest.Init(t, s)

	// Title is empty; ctrl+s must not dismiss or create a task.
	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.False(t, dismissed, "overlay must not dismiss with empty title")

	tasks, err := svc.ListTasks(t.Context(), gtd.TaskFilter{})
	require.NoError(t, err)
	assert.Empty(t, tasks, "no task should be created")
}
