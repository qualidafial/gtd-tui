package taskedit

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/qualidafial/gtd-tui"
)

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
			m := New(gtd.Task{ID: 1, Title: "Existing", Status: tt.status, StatusChangedAt: tt.at}, nil)
			view := ansi.Strip(m.View())
			if !strings.Contains(view, tt.want) {
				t.Fatalf("expected status line %q in view, got:\n%s", tt.want, view)
			}
		})
	}
}

func TestModel_SaveError_RendersInView(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil)

	updated, _ := m.Update(taskSavedMsg{err: errors.New("disk full")})
	view := updated.(Model).View()

	if !strings.Contains(view, "disk full") {
		t.Fatalf("expected save error in view, got:\n%s", view)
	}
}

func TestModel_SaveError_EscClearsErrorAndResumesForm(t *testing.T) {
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil)

	withErr, _ := m.Update(taskSavedMsg{err: errors.New("disk full")})
	if withErr.(Model).err == nil {
		t.Fatal("precondition: expected error to be set after save failure")
	}

	cleared, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		t.Fatalf("expected no cmd from esc-clear; got %v", cmd)
	}
	if cleared.(Model).err != nil {
		t.Fatalf("expected err to be cleared after esc; got %v", cleared.(Model).err)
	}
	if cleared.(Model).form.State != huh.StateNormal {
		t.Fatalf("expected form state reset to StateNormal so user can retry; got %v", cleared.(Model).form.State)
	}
}

func TestModel_SaveError_OtherKeysSwallowed(t *testing.T) {
	// Regression guard: after a save error the form is stuck in
	// StateCompleted, so any keypress that fell through to the form would
	// re-fire the save and spin in a loop. Non-esc keys must be swallowed.
	m := New(gtd.Task{ID: 1, Title: "Existing"}, nil)

	withErr, _ := m.Update(taskSavedMsg{err: errors.New("disk full")})

	_, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if cmd != nil {
		t.Fatalf("expected nil cmd on non-esc key after save error, got %v", cmd)
	}
}
