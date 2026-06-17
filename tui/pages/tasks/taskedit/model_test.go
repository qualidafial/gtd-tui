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

func TestCreate_StatusRadio_DefaultsOpenOffersDoneNotDropped(t *testing.T) {
	var s screen.Screen = New(gtd.Task{}, nil, "", nil)
	s = screentest.Init(t, s)
	s = screentest.Send(t, s, tea.WindowSizeMsg{Width: 80, Height: 24})

	view := ansi.Strip(s.View())
	assert.Contains(t, view, "Status")
	assert.Contains(t, view, "Open")
	assert.Contains(t, view, "Done")
	assert.NotContains(t, view, "Dropped", "new tasks must not offer Dropped")

	assert.Equal(t, gtd.TaskStatusOpen, s.(Model).form.FieldValues()["status"],
		"the status radio defaults to Open")
}

func TestCreate_OpenStatus_CreatesOpenTask(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewTaskService(db)

	var s screen.Screen = New(gtd.Task{}, svc, "", nil)
	s = screentest.Init(t, s)
	s = screentest.TypeText(t, s, "Recorded task")
	// Title→Description→Assignee→Due→Defer→Status.
	for i := 0; i < 5; i++ {
		s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyTab})
	}
	require.Equal(t, "status", s.(Model).form.Focused().Key())

	// Default selection is Open; Enter on the terminal radio submits.
	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, dismissed)

	tasks, err := svc.ListTasks(t.Context(), gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Recorded task", tasks[0].Title)
	assert.Equal(t, gtd.TaskStatusOpen, tasks[0].Status)
}

func TestCreate_DoneStatus_CreatesDoneTask(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewTaskService(db)

	var s screen.Screen = New(gtd.Task{}, svc, "", nil)
	s = screentest.Init(t, s)
	s = screentest.TypeText(t, s, "Already finished")
	for i := 0; i < 5; i++ {
		s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyTab})
	}
	require.Equal(t, "status", s.(Model).form.Focused().Key())

	// Move the selection Open→Done, then submit on the terminal radio.
	s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyRight})
	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, dismissed)

	tasks, err := svc.ListTasks(t.Context(), gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Already finished", tasks[0].Title)
	assert.Equal(t, gtd.TaskStatusDone, tasks[0].Status)
	assert.WithinDuration(t, time.Now(), tasks[0].StatusChangedAt, time.Minute,
		"a recorded done task is stamped done at creation")
}

func TestEdit_NoStatusField(t *testing.T) {
	var s screen.Screen = New(gtd.Task{ID: 1, Title: "Existing", Status: gtd.TaskStatusOpen}, nil, "", nil)
	s = screentest.Init(t, s)
	s = screentest.Send(t, s, tea.WindowSizeMsg{Width: 80, Height: 24})

	_, ok := s.(Model).form.FieldValues()["status"]
	assert.False(t, ok, "existing-task form must not expose a status field")
	assert.Contains(t, ansi.Strip(s.View()), "Save", "existing-task form keeps the Save button")
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
