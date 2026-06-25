package tasklist

import (
	"context"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskedit"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

// viewStub is a no-op Screen returned by a test view factory.
type viewStub struct{}

func (viewStub) Init() tea.Cmd                             { return nil }
func (s viewStub) Update(tea.Msg) (screen.Screen, tea.Cmd) { return s, nil }
func (viewStub) View() string                              { return "" }
func (viewStub) Keys() []keymap.Group                      { return nil }

// pushedScreen runs cmd, asserts it yields a PushMsg, and returns the pushed
// screen.
func pushedScreen(t *testing.T, cmd tea.Cmd) screen.Screen {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a cmd")
	}
	push, ok := cmd().(screen.PushMsg)
	if !ok {
		t.Fatalf("expected PushMsg, got %T", cmd())
	}
	return push.Screen
}

// selectOne builds a tasklist with the given view factory and a single
// selected open task.
func selectOne(viewFn ViewFactory) Model {
	m := New(nil, "", nil, nil, nil, false, viewFn)
	loaded, _ := m.Update(TasksLoadedMsg{
		tasks: []gtd.Task{{ID: 5, Title: "t", Status: gtd.TaskStatusOpen}},
	})
	return loaded.(Model)
}

func TestModel_EnterKey_OpensView(t *testing.T) {
	var got gtd.Task
	m := selectOne(func(task gtd.Task) screen.Screen {
		got = task
		return viewStub{}
	})

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if _, ok := pushedScreen(t, cmd).(viewStub); !ok {
		t.Fatal("enter should push the task view")
	}
	if got.ID != 5 {
		t.Fatalf("view factory received task %d, want 5", got.ID)
	}
}

func TestModel_EnterKey_FallsBackToEditorWithoutViewFn(t *testing.T) {
	m := selectOne(nil)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if _, ok := pushedScreen(t, cmd).(taskedit.Model); !ok {
		t.Fatal("enter without a view factory should open the editor")
	}
}

func TestModel_EKey_OpensEditor(t *testing.T) {
	// A view factory is present to prove "e" still edits rather than viewing.
	m := selectOne(func(gtd.Task) screen.Screen { return viewStub{} })

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})

	if _, ok := pushedScreen(t, cmd).(taskedit.Model); !ok {
		t.Fatal("e should open the task editor")
	}
}

func openTestTaskSvc(t *testing.T) gtd.TaskService {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return service.NewTaskService(db)
}

// applied returns a tasklist whose query bar holds a non-default applied query,
// with the reset binding reconciled as it would be after a reload.
func applied(svc gtd.TaskService, defaultQuery, applied string) Model {
	m := New(svc, defaultQuery, nil, nil, nil, false, nil)
	m.query.SetValue(applied)
	m.updateKeybindings()
	return m
}

func TestModel_ResetQuery_RevertsToDefaultAndReloads(t *testing.T) {
	m := applied(openTestTaskSvc(t), "status:open ready:now", "status:done")

	updated, cmd := m.Update(tea.KeyPressMsg{Code: '\\', Text: "\\"})
	got := updated.(Model)
	if v := got.query.Value(); v != "status:open ready:now" {
		t.Fatalf("query after revert = %q, want default", v)
	}
	if cmd == nil {
		t.Fatal("expected a reload cmd from revert")
	}
	if _, ok := cmd().(TasksLoadedMsg); !ok {
		t.Fatal("revert cmd should reload tasks")
	}
}

func TestModel_ResetQuery_EscAfterRevertRevertsToDefault(t *testing.T) {
	m := applied(openTestTaskSvc(t), "status:open ready:now", "status:done")

	reverted, _ := m.Update(tea.KeyPressMsg{Code: '\\', Text: "\\"})
	// Focus then esc: esc reverts to the applied query, proving the revert
	// recorded the default as the applied query.
	focused, _ := reverted.(Model).Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	canceled, _ := focused.(Model).Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if v := canceled.(Model).query.Value(); v != "status:open ready:now" {
		t.Fatalf("query after esc = %q, want default", v)
	}
}

func TestModel_ResetQuery_DisabledAtDefault(t *testing.T) {
	m := New(openTestTaskSvc(t), "status:open ready:now", nil, nil, nil, false, nil)
	if m.KeyMap.ResetQuery.Enabled() {
		t.Fatal("ResetQuery should be disabled when query equals default")
	}
	// Pressing it is a no-op: no reload cmd, value unchanged.
	updated, cmd := m.Update(tea.KeyPressMsg{Code: '\\', Text: "\\"})
	if cmd != nil {
		t.Fatal("revert at default should not issue a cmd")
	}
	if v := updated.(Model).query.Value(); v != "status:open ready:now" {
		t.Fatalf("value changed at default: %q", v)
	}
}

func TestModel_ResetQuery_InertWhileEditing(t *testing.T) {
	m := New(openTestTaskSvc(t), "status:open ready:now", nil, nil, nil, false, nil)
	focused, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	typed, _ := focused.(Model).Update(tea.KeyPressMsg{Code: '\\', Text: "\\"})
	got := typed.(Model)
	if !got.CapturingInput() {
		t.Fatal("query bar should still be focused after typing backslash")
	}
	if v := got.query.Value(); v != "status:open ready:now \\" {
		t.Fatalf("backslash not entered into query: %q", v)
	}
}

func TestModel_ResetQuery_EmptyDefaultClears(t *testing.T) {
	m := applied(openTestTaskSvc(t), "", "status:done")

	updated, cmd := m.Update(tea.KeyPressMsg{Code: '\\', Text: "\\"})
	if v := updated.(Model).query.Value(); v != "" {
		t.Fatalf("query after revert = %q, want empty", v)
	}
	if cmd == nil {
		t.Fatal("expected a reload cmd")
	}
}

func TestModel_TasksLoaded_AppliesItems(t *testing.T) {
	var svc gtd.TaskService = nil
	pending := New(svc, "status:open", nil, nil, nil, false, nil)

	updated, _ := pending.Update(TasksLoadedMsg{
		tasks: []gtd.Task{{ID: 1, Title: "open task", Status: gtd.TaskStatusOpen}},
	})

	got := updated.(Model).list.Items()
	if len(got) != 1 {
		t.Fatalf("expected 1 item; got %d", len(got))
	}
}

func TestModel_NewTaskKey_OpensEditor(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{name: "c", key: tea.KeyPressMsg{Code: 'c', Text: "c"}},
		{name: "insert", key: tea.KeyPressMsg{Code: tea.KeyInsert}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(nil, "status:open", nil, nil, nil, false, nil)
			_, cmd := m.Update(tt.key)
			if cmd == nil {
				t.Fatal("expected a cmd from new-task keybinding")
			}
			msg := cmd()
			if _, ok := msg.(screen.PushMsg); !ok {
				t.Fatalf("expected PushMsg, got %T", msg)
			}
		})
	}
}

func TestModel_NKey_NoLongerOpensEditor(t *testing.T) {
	m := New(nil, "status:open", nil, nil, nil, false, nil)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, ok := msg.(screen.PushMsg); ok {
				t.Fatal("'n' should no longer trigger the new-task overlay")
			}
		}
	}
}

// loadOne builds a tasklist with a single selected task of the given status.
func loadOne(status gtd.TaskStatus) Model {
	m := New(nil, "", nil, nil, nil, false, nil)
	updated, _ := m.Update(TasksLoadedMsg{
		tasks: []gtd.Task{{ID: 1, Title: "t", Status: status}},
	})
	return updated.(Model)
}

func overlayTransition(t *testing.T, cmd tea.Cmd) (taskstatus.Transition, bool) {
	t.Helper()
	if cmd == nil {
		return 0, false
	}
	msg := cmd()
	push, ok := msg.(screen.PushMsg)
	if !ok {
		return 0, false
	}
	ov, ok := push.Screen.(taskstatus.Model)
	if !ok {
		t.Fatalf("overlay is %T, want taskstatus.Model", push.Screen)
	}
	return ov.Transition(), true
}

func TestModel_Status_PushesPickerSeeded(t *testing.T) {
	for _, status := range []gtd.TaskStatus{gtd.TaskStatusOpen, gtd.TaskStatusDone, gtd.TaskStatusDropped} {
		t.Run(string(status), func(t *testing.T) {
			m := loadOne(status)
			_, cmd := m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
			ov, ok := pushedScreen(t, cmd).(taskstatus.Model)
			if !ok {
				t.Fatalf("s should push the status picker, got %T", pushedScreen(t, cmd))
			}
			if !ov.Picking() {
				t.Fatal("s should open the picker, not a fixed-transition confirmation")
			}
			if ov.Current() != status {
				t.Fatalf("picker seeded with %v, want %v", ov.Current(), status)
			}
		})
	}
}

func TestModel_Delete_DropsPendingTask(t *testing.T) {
	m := loadOne(gtd.TaskStatusOpen)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	got, ok := overlayTransition(t, cmd)
	if !ok {
		t.Fatal("delete did not open a transition overlay")
	}
	if got != taskstatus.Drop {
		t.Fatalf("transition = %v, want Drop", got)
	}
}

func TestModel_Delete_NoOpOnClosedTasks(t *testing.T) {
	for _, status := range []gtd.TaskStatus{gtd.TaskStatusDone, gtd.TaskStatusDropped} {
		t.Run(string(status), func(t *testing.T) {
			m := loadOne(status)
			_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
			if _, ok := overlayTransition(t, cmd); ok {
				t.Fatalf("delete on a %s task should not open an overlay", status)
			}
		})
	}
}

func TestModel_MoveBindings_Boundaries(t *testing.T) {
	m := New(nil, "", nil, nil, nil, false, nil)
	upd, _ := m.Update(TasksLoadedMsg{
		tasks: []gtd.Task{
			{ID: 1, Title: "a", Status: gtd.TaskStatusOpen},
			{ID: 2, Title: "b", Status: gtd.TaskStatusOpen},
			{ID: 3, Title: "c", Status: gtd.TaskStatusDone},
		},
	})
	m = upd.(Model)

	down := func(m Model) Model {
		u, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		return u.(Model)
	}

	tests := []struct {
		name           string
		model          Model
		wantUp, wantDn bool
	}{
		{"first pending", m, false, true},
		{"last pending", down(m), true, false},
		{"done task", down(down(m)), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.KeyMap.MoveUp.Enabled(); got != tt.wantUp {
				t.Errorf("MoveUp enabled = %v, want %v", got, tt.wantUp)
			}
			if got := tt.model.KeyMap.MoveDown.Enabled(); got != tt.wantDn {
				t.Errorf("MoveDown enabled = %v, want %v", got, tt.wantDn)
			}
			// Move-to-top tracks move-up; move-to-bottom tracks move-down.
			if got := tt.model.KeyMap.MoveFirst.Enabled(); got != tt.wantUp {
				t.Errorf("MoveFirst enabled = %v, want %v", got, tt.wantUp)
			}
			if got := tt.model.KeyMap.MoveLast.Enabled(); got != tt.wantDn {
				t.Errorf("MoveLast enabled = %v, want %v", got, tt.wantDn)
			}
		})
	}
}

func TestModel_MoveLast_ReordersAndKeepsCursor(t *testing.T) {
	svc := openTestTaskSvc(t)
	ctx := context.Background()

	var first gtd.Task
	for _, title := range []string{"a", "b", "c"} {
		task, err := svc.CreateTask(ctx, gtd.Task{Title: title, Status: gtd.TaskStatusOpen})
		if err != nil {
			t.Fatalf("create task: %v", err)
		}
		if title == "a" {
			first = task
		}
	}

	tasks, err := svc.ListTasks(ctx, gtd.TaskFilter{})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}

	m := New(svc, "", nil, nil, nil, false, nil)
	loaded, _ := m.Update(TasksLoadedMsg{tasks: tasks})
	m = loaded.(Model)

	// Cursor starts on the first task (a). shift+end moves it to the bottom.
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnd, Mod: tea.ModShift})
	if cmd == nil {
		t.Fatal("expected a reorder cmd from shift+end")
	}
	msg, ok := cmd().(tasksReorderedMsg)
	if !ok {
		t.Fatalf("expected tasksReorderedMsg, got %T", cmd())
	}
	if msg.selectID != first.ID {
		t.Fatalf("selectID = %d, want %d (the moved task)", msg.selectID, first.ID)
	}
	gotOrder := make([]string, len(msg.tasks))
	for i, task := range msg.tasks {
		gotOrder[i] = task.Title
	}
	if want := []string{"b", "c", "a"}; !slices.Equal(gotOrder, want) {
		t.Fatalf("order after move = %v, want %v", gotOrder, want)
	}

	// Applying the msg keeps the cursor on the moved task.
	applied, _ := m.Update(msg)
	if sel, ok := applied.(Model).list.SelectedItem().(Item); !ok || sel.task.ID != first.ID {
		t.Fatal("cursor not on moved task after reorder")
	}
}
