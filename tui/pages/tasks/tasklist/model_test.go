package tasklist

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

// TestModel_TasksLoaded_IgnoresOtherFilter guards the startup bug where both
// tasklist tabs receive every TasksLoadedMsg broadcast: a tab must ignore loads
// addressed to a different filter so results don't overwrite the wrong tab.
func TestModel_TasksLoaded_IgnoresOtherFilter(t *testing.T) {
	var svc gtd.TaskService = nil
	pending := New(svc, "status:pending")

	updated, _ := pending.Update(TasksLoadedMsg{
		filter: gtd.TaskFilter{}.WithStatus(gtd.TaskStatusDone),
		tasks:  []gtd.Task{{ID: 1, Title: "done task", Status: gtd.TaskStatusDone}},
	})

	got := updated.(Model).list.Items()
	if len(got) != 0 {
		t.Fatalf("expected no items when filter does not match; got %d", len(got))
	}
}

func TestModel_TasksLoaded_AppliesMatchingFilter(t *testing.T) {
	var svc gtd.TaskService = nil
	pending := New(svc, "status:pending")

	updated, _ := pending.Update(TasksLoadedMsg{
		filter: gtd.TaskFilter{}.WithStatus(gtd.TaskStatusPending),
		tasks:  []gtd.Task{{ID: 1, Title: "pending task", Status: gtd.TaskStatusPending}},
	})

	got := updated.(Model).list.Items()
	if len(got) != 1 {
		t.Fatalf("expected 1 item when filter matches; got %d", len(got))
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
			m := New(nil, "status:pending")
			_, cmd := m.Update(tt.key)
			if cmd == nil {
				t.Fatal("expected a cmd from new-task keybinding")
			}
			msg := cmd()
			if _, ok := msg.(screen.ShowOverlayMsg); !ok {
				t.Fatalf("expected ShowOverlayMsg, got %T", msg)
			}
		})
	}
}

func TestModel_NKey_NoLongerOpensEditor(t *testing.T) {
	m := New(nil, "status:pending")
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, ok := msg.(screen.ShowOverlayMsg); ok {
				t.Fatal("'n' should no longer trigger the new-task overlay")
			}
		}
	}
}

// loadOne builds a tasklist with a single selected task of the given status.
func loadOne(status gtd.TaskStatus) Model {
	m := New(nil, "")
	updated, _ := m.Update(TasksLoadedMsg{
		filter: gtd.TaskFilter{},
		tasks:  []gtd.Task{{ID: 1, Title: "t", Status: status}},
	})
	return updated.(Model)
}

func overlayTransition(t *testing.T, cmd tea.Cmd) (taskstatus.Transition, bool) {
	t.Helper()
	if cmd == nil {
		return 0, false
	}
	msg := cmd()
	show, ok := msg.(screen.ShowOverlayMsg)
	if !ok {
		return 0, false
	}
	ov, ok := show.Overlay.(taskstatus.Model)
	if !ok {
		t.Fatalf("overlay is %T, want taskstatus.Model", show.Overlay)
	}
	return ov.Transition(), true
}

func TestModel_Toggle_ResolvesTransition(t *testing.T) {
	tests := []struct {
		status gtd.TaskStatus
		want   taskstatus.Transition
	}{
		{gtd.TaskStatusPending, taskstatus.Complete},
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
	m := loadOne(gtd.TaskStatusPending)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	got, ok := overlayTransition(t, cmd)
	if !ok {
		t.Fatal("delete did not open a transition overlay")
	}
	if got != taskstatus.Drop {
		t.Fatalf("transition = %v, want Drop", got)
	}
}

// Drop is only valid from pending: the service rejects dropping done or dropped
// tasks, so the binding is disabled and delete is inert for both.
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
	// pending, pending, done — pending tasks sort above the closed one.
	m := New(nil, "")
	upd, _ := m.Update(TasksLoadedMsg{
		filter: gtd.TaskFilter{},
		tasks: []gtd.Task{
			{ID: 1, Title: "a", Status: gtd.TaskStatusPending},
			{ID: 2, Title: "b", Status: gtd.TaskStatusPending},
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

func TestFilterMatches(t *testing.T) {
	tests := []struct {
		name string
		a, b gtd.TaskFilter
		want bool
	}{
		{
			name: "same single status",
			a:    gtd.TaskFilter{}.WithStatus(gtd.TaskStatusPending),
			b:    gtd.TaskFilter{}.WithStatus(gtd.TaskStatusPending),
			want: true,
		},
		{
			name: "different status",
			a:    gtd.TaskFilter{}.WithStatus(gtd.TaskStatusPending),
			b:    gtd.TaskFilter{}.WithStatus(gtd.TaskStatusDone),
			want: false,
		},
		{
			name: "different task ids",
			a:    gtd.TaskFilter{}.WithTaskIDs(1),
			b:    gtd.TaskFilter{}.WithTaskIDs(2),
			want: false,
		},
		{
			name: "both empty",
			a:    gtd.TaskFilter{},
			b:    gtd.TaskFilter{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterMatches(tt.a, tt.b); got != tt.want {
				t.Errorf("filterMatches(%+v, %+v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
