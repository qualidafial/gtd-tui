package taskedit

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/qualidafial/gtd-tui"
)

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
