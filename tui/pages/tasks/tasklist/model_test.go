package tasklist

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

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
			if got := tt.model.keys.MoveUp.Enabled(); got != tt.wantUp {
				t.Errorf("MoveUp enabled = %v, want %v", got, tt.wantUp)
			}
			if got := tt.model.keys.MoveDown.Enabled(); got != tt.wantDn {
				t.Errorf("MoveDown enabled = %v, want %v", got, tt.wantDn)
			}
		})
	}
}