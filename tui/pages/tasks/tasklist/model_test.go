package tasklist

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
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
