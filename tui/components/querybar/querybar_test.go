package querybar

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func alwaysValid(_ string) *ParseError { return nil }

func failOn(bad string) func(string) *ParseError {
	return func(s string) *ParseError {
		if s == bad {
			return &ParseError{Message: "bad value", Start: 0, End: len([]rune(s))}
		}
		return nil
	}
}

func newFocused(t *testing.T, query string) Model {
	t.Helper()
	m := New("/ ", "(all)", 2*time.Second, alwaysValid)
	m.SetValue(query)
	m2, _ := m.Focus()
	return m2
}

func TestFocus_AppendSpace_NonEmpty(t *testing.T) {
	m := New("/ ", "", 2*time.Second, alwaysValid)
	m.SetValue("status:open")
	m2, _ := m.Focus()
	if got := m2.Value(); got != "status:open " {
		t.Fatalf("Focus() value = %q, want %q", got, "status:open ")
	}
	if !m2.CapturingInput() {
		t.Fatal("CapturingInput() should be true after Focus()")
	}
}

func TestFocus_EmptyValue_NoSpace(t *testing.T) {
	m := New("/ ", "", 2*time.Second, alwaysValid)
	m2, _ := m.Focus()
	if got := m2.Value(); got != "" {
		t.Fatalf("Focus() on empty should not add space, got %q", got)
	}
}

func TestBlur_TrimsWhitespace(t *testing.T) {
	m := New("/ ", "", 2*time.Second, alwaysValid)
	m.SetValue("  foo  ")
	m2, _ := m.Focus()
	m3 := m2.Blur()
	if got := m3.Value(); got != "foo" {
		t.Fatalf("Blur() value = %q, want %q", got, "foo")
	}
	if m3.CapturingInput() {
		t.Fatal("CapturingInput() should be false after Blur()")
	}
}

func TestApply_Success(t *testing.T) {
	m := newFocused(t, "status:open")
	m2, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from apply")
	}
	msg := cmd()
	apply, ok := msg.(ApplyMsg)
	if !ok {
		t.Fatalf("expected ApplyMsg, got %T", msg)
	}
	if apply.Query != "status:open" {
		t.Fatalf("ApplyMsg.Query = %q, want %q", apply.Query, "status:open")
	}
	if m2.CapturingInput() {
		t.Fatal("should be blurred after successful apply")
	}
}

func TestApply_Failure_ReturnsError(t *testing.T) {
	m := New("/ ", "", 2*time.Second, failOn("bad"))
	m.SetValue("bad")
	m2, _ := m.Focus()
	m3, cmd := m2.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from failed apply")
	}
	msg := cmd()
	if _, ok := msg.(*ParseError); !ok {
		t.Fatalf("expected *ParseError error msg, got %T", msg)
	}
	if !m3.CapturingInput() {
		t.Fatal("should remain focused after failed apply")
	}
}

func TestCancel_RevertsAndRevertApplies(t *testing.T) {
	// Seed an applied query, focus (which appends a trailing space), then type
	// extra characters so the working value diverges from the applied query.
	m := New("/ ", "", 2*time.Second, alwaysValid)
	m.SetValue("status:open")
	m2, _ := m.Focus()
	m2, _ = m2.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})

	m3, cmd := m2.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected cmd from esc")
	}
	msg := cmd()
	apply, ok := msg.(ApplyMsg)
	if !ok {
		t.Fatalf("expected ApplyMsg on esc, got %T", msg)
	}
	if apply.Query != "status:open" {
		t.Fatalf("esc ApplyMsg.Query = %q, want previously-applied %q", apply.Query, "status:open")
	}
	if m3.CapturingInput() {
		t.Fatal("should be blurred after esc")
	}
	if got := m3.Value(); got != "status:open" {
		t.Fatalf("esc value = %q, want reverted to %q", got, "status:open")
	}
}

func TestDebounce_IgnoresStaleSeq(t *testing.T) {
	m := New("/ ", "", 10*time.Millisecond, failOn("bad"))
	m.SetValue("bad")
	m2, _ := m.Focus()
	// Bump the seq with a real keystroke so seq 0 is genuinely stale.
	m2, _ = m2.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})

	m3, cmd := m2.Update(debounceMsg{seq: 0})
	if m3.parseErr != nil {
		t.Fatal("stale debounce should not update parseErr")
	}
	if cmd != nil {
		t.Fatal("stale debounce should not emit any cmd")
	}
}

func TestDebounce_InvalidQuery_EmitsParseError(t *testing.T) {
	// Validate fails on any non-empty input so Focus()'s trailing space
	// (trimmed inside the debounce handler) doesn't confuse the test.
	validate := func(s string) *ParseError {
		if len(s) > 0 {
			return &ParseError{Message: "bad", Start: 0, End: len([]rune(s))}
		}
		return nil
	}
	m := New("/ ", "", 10*time.Millisecond, validate)
	m.SetValue("bad")
	m2, _ := m.Focus()
	seq := m2.debounceSeq
	m3, cmd := m2.Update(debounceMsg{seq: seq})
	if m3.parseErr == nil {
		t.Fatal("debounce should have set parseErr")
	}
	if cmd == nil {
		t.Fatal("debounce should return a cmd on parse failure")
	}
	if _, ok := cmd().(*ParseError); !ok {
		t.Fatalf("expected *ParseError from debounce cmd, got %T", cmd())
	}
	if !m3.CapturingInput() {
		t.Fatal("debounce should not blur on parse failure")
	}
}

func TestDebounce_ValidQuery_EmitsApplyButStaysFocused(t *testing.T) {
	m := New("/ ", "", 10*time.Millisecond, alwaysValid)
	m.SetValue("status:open")
	m2, _ := m.Focus()
	// Focus appends a trailing space; appliedQuery stays "status:open".
	appliedBefore := m2.appliedQuery

	seq := m2.debounceSeq
	m3, cmd := m2.Update(debounceMsg{seq: seq})
	if cmd == nil {
		t.Fatal("expected ApplyMsg cmd from successful debounced parse")
	}
	apply, ok := cmd().(ApplyMsg)
	if !ok {
		t.Fatalf("expected ApplyMsg, got %T", cmd())
	}
	if apply.Query != "status:open" {
		t.Fatalf("ApplyMsg.Query = %q, want trimmed %q", apply.Query, "status:open")
	}
	if !m3.CapturingInput() {
		t.Fatal("query bar should remain focused after live-preview apply")
	}
	if m3.appliedQuery != appliedBefore {
		t.Fatalf("appliedQuery should not change on debounce apply: was %q, now %q",
			appliedBefore, m3.appliedQuery)
	}
	if m3.parseErr != nil {
		t.Fatalf("parseErr should be cleared on successful debounced parse, got %v", m3.parseErr)
	}
}

func TestApply_TrimsWhitespace(t *testing.T) {
	m := newFocused(t, "status:open")
	// value has trailing space from Focus
	m2, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_ = m2
	msg := cmd()
	apply, ok := msg.(ApplyMsg)
	if !ok {
		t.Fatalf("expected ApplyMsg, got %T", msg)
	}
	if apply.Query != "status:open" {
		t.Fatalf("apply should trim whitespace, got %q", apply.Query)
	}
}
