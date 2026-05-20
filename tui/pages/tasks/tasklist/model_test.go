package tasklist

import (
	"testing"

	"github.com/qualidafial/gtd-tui"
)

// TestModel_TasksLoaded_IgnoresOtherFilter guards the startup bug where both
// tasklist tabs receive every TasksLoadedMsg broadcast: a tab must ignore loads
// addressed to a different filter so Active-tab results don't overwrite the
// Inbox tab's items (and vice versa).
func TestModel_TasksLoaded_IgnoresOtherFilter(t *testing.T) {
	var svc gtd.TaskService = nil
	inbox := New(svc, gtd.TaskFilter{}.Status(gtd.TaskStatusInbox))

	updated, _ := inbox.Update(TasksLoadedMsg{
		filter: gtd.TaskFilter{}.Status(gtd.TaskStatusActive),
		tasks:  []gtd.Task{{ID: 1, Title: "active task", Status: gtd.TaskStatusActive}},
	})

	got := updated.(Model).list.Items()
	if len(got) != 0 {
		t.Fatalf("expected no items when filter does not match; got %d", len(got))
	}
}

func TestModel_TasksLoaded_AppliesMatchingFilter(t *testing.T) {
	var svc gtd.TaskService = nil
	inbox := New(svc, gtd.TaskFilter{}.Status(gtd.TaskStatusInbox))

	updated, _ := inbox.Update(TasksLoadedMsg{
		filter: gtd.TaskFilter{}.Status(gtd.TaskStatusInbox),
		tasks:  []gtd.Task{{ID: 1, Title: "inbox task", Status: gtd.TaskStatusInbox}},
	})

	got := updated.(Model).list.Items()
	if len(got) != 1 {
		t.Fatalf("expected 1 item when filter matches; got %d", len(got))
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
			a:    gtd.TaskFilter{}.Status(gtd.TaskStatusInbox),
			b:    gtd.TaskFilter{}.Status(gtd.TaskStatusInbox),
			want: true,
		},
		{
			name: "different status",
			a:    gtd.TaskFilter{}.Status(gtd.TaskStatusInbox),
			b:    gtd.TaskFilter{}.Status(gtd.TaskStatusActive),
			want: false,
		},
		{
			name: "different task ids",
			a:    gtd.TaskFilter{}.TaskID(1),
			b:    gtd.TaskFilter{}.TaskID(2),
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
