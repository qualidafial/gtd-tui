package tasklist

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

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
	m := New(svc, defaultQuery, nil, nil, false)
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
	m := New(openTestTaskSvc(t), "status:open ready:now", nil, nil, false)
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
	m := New(openTestTaskSvc(t), "status:open ready:now", nil, nil, false)
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
	pending := New(svc, "status:open", nil, nil, false)

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
		{name: "plus", key: tea.KeyPressMsg{Code: '+', Text: "+"}},
		{name: "insert", key: tea.KeyPressMsg{Code: tea.KeyInsert}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(nil, "status:open", nil, nil, false)
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
	m := New(nil, "status:open", nil, nil, false)
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
	m := New(nil, "", nil, nil, false)
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

func TestModel_Toggle_ResolvesTransition(t *testing.T) {
	tests := []struct {
		status gtd.TaskStatus
		want   taskstatus.Transition
	}{
		{gtd.TaskStatusOpen, taskstatus.Complete},
		{gtd.TaskStatusDone, taskstatus.Reopen},
		{gtd.TaskStatusDropped, taskstatus.Reopen},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			m := loadOne(tt.status)
			_, cmd := m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
			got, ok := overlayTransition(t, cmd)
			if !ok {
				t.Fatal("space did not open a transition overlay")
			}
			if got != tt.want {
				t.Fatalf("transition = %v, want %v", got, tt.want)
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
	m := New(nil, "", nil, nil, false)
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
		})
	}
}
