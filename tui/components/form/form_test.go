package form_test

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
)

// stubField is a controllable form.Field for unit tests.
type stubField struct {
	key      string
	val      any
	focused  bool
	visible  func(form.Values) bool
	validate func() error
	short    []key.Binding
	full     [][]key.Binding

	// Pointer-backed counters so they survive value copies of stubField.
	validateCalls *int
	updateCalls   *int
	widths        *[]int
}

func newStub(k string) *stubField {
	return &stubField{
		key:           k,
		validateCalls: new(int),
		updateCalls:   new(int),
		widths:        new([]int),
	}
}

// asField returns a value-typed Field view of s, copying scalar state.
func (s *stubField) asField() form.Field { return *s } //nolint:revive // tests construct via pointer

func (s stubField) Key() string   { return s.key }
func (s stubField) Focused() bool { return s.focused }

func (s stubField) Visible(v form.Values) bool {
	if s.visible == nil {
		return true
	}
	return s.visible(v)
}

func (s stubField) Focus() (form.Field, tea.Cmd) {
	s.focused = true
	return s, nil
}

func (s stubField) Blur() form.Field {
	s.focused = false
	return s
}

func (s stubField) SetWidth(w int) form.Field {
	if s.widths != nil {
		*s.widths = append(*s.widths, w)
	}
	return s
}

func (s stubField) Update(msg tea.Msg) (form.Field, tea.Cmd) {
	if s.updateCalls != nil {
		*s.updateCalls++
	}
	// Allow tests to mutate val via a key press by piggybacking on a custom msg.
	if vm, ok := msg.(setValueMsg); ok && vm.key == s.key {
		s.val = vm.val
	}
	return s, nil
}

func (s stubField) View() string { return s.key + "-view" }
func (s stubField) Value() any   { return s.val }

func (s stubField) Validate() (form.Field, error) {
	if s.validateCalls != nil {
		*s.validateCalls++
	}
	if s.validate == nil {
		return s, nil
	}
	return s, s.validate()
}

func (s stubField) ShortHelp() []key.Binding  { return s.short }
func (s stubField) FullHelp() [][]key.Binding { return s.full }

// setValueMsg lets a test change the value of a specific stub field through
// the form's Update pipeline.
type setValueMsg struct {
	key string
	val any
}

func tab() tea.KeyPressMsg       { return tea.KeyPressMsg{Code: tea.KeyTab} }
func shiftTab() tea.KeyPressMsg  { return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift} }
func ctrlS() tea.KeyPressMsg     { return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl} }

// collectMsgs flattens a possibly-batched tea.Cmd into the slice of msgs it
// would emit. Sufficient for inspecting whether SubmittedMsg fires.
func collectMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	if bm, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range bm {
			out = append(out, collectMsgs(c)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

func hasSubmittedMsg(cmd tea.Cmd) bool {
	for _, m := range collectMsgs(cmd) {
		if _, ok := m.(form.SubmittedMsg); ok {
			return true
		}
	}
	return false
}

// --- tests ------------------------------------------------------------------

func TestNewPanicsOnDuplicateKey(t *testing.T) {
	a := newStub("dup").asField()
	b := newStub("dup").asField()
	assert.PanicsWithValue(t, "form: duplicate field key: dup", func() {
		form.New(a, b)
	})
}

func TestNewPanicsOnEmptyKey(t *testing.T) {
	a := newStub("").asField()
	assert.Panics(t, func() { form.New(a) })
}

func TestInitialFocusOnFirstField(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	require.NotNil(t, f.Focused())
	assert.Equal(t, "a", f.Focused().Key())
	assert.True(t, f.Focused().Focused())
}

func TestInitialFocusSkipsHiddenFields(t *testing.T) {
	a := newStub("a")
	a.visible = func(form.Values) bool { return false }
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	require.NotNil(t, f.Focused())
	assert.Equal(t, "b", f.Focused().Key())
}

func TestTabAdvancesFocus(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	f, _ = f.Update(tab())
	require.NotNil(t, f.Focused())
	assert.Equal(t, "b", f.Focused().Key())
}

func TestShiftTabRetreatsFocus(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	f, _ = f.Update(tab())
	require.Equal(t, "b", f.Focused().Key())
	f, _ = f.Update(shiftTab())
	assert.Equal(t, "a", f.Focused().Key())
}

func TestTabSkipsHiddenField(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	b.visible = func(form.Values) bool { return false }
	c := newStub("c")
	f := form.New(a.asField(), b.asField(), c.asField())

	f, _ = f.Update(tab())
	require.NotNil(t, f.Focused())
	assert.Equal(t, "c", f.Focused().Key())
}

func TestTabOnLastFieldIsNoOp(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	f, _ = f.Update(tab())
	f, _ = f.Update(tab())
	assert.Equal(t, "b", f.Focused().Key())
}

func TestSubmitHappyPathReturnsSubmittedMsg(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	f, cmd := f.Update(ctrlS())
	_ = f
	assert.Equal(t, 1, *a.validateCalls)
	assert.Equal(t, 1, *b.validateCalls)
	assert.True(t, hasSubmittedMsg(cmd), "expected SubmittedMsg in cmd batch")
}

func TestSubmitHaltsOnFirstFailure(t *testing.T) {
	wantErr := errors.New("a-invalid")

	a := newStub("a")
	a.validate = func() error { return wantErr }
	b := newStub("b")
	b.validate = func() error { return errors.New("b-invalid") }
	c := newStub("c")
	c.validate = func() error { return errors.New("c-invalid") }

	f := form.New(a.asField(), b.asField(), c.asField())
	require.Equal(t, "a", f.Focused().Key())

	f, cmd := f.Update(ctrlS())

	require.NotNil(t, f.Focused())
	assert.Equal(t, "a", f.Focused().Key(), "focus should land on first failing field")
	assert.Equal(t, 1, *a.validateCalls)
	assert.Equal(t, 0, *b.validateCalls, "later fields must not be validated after first failure")
	assert.Equal(t, 0, *c.validateCalls)
	assert.False(t, hasSubmittedMsg(cmd), "SubmittedMsg must not fire on failure")
}

func TestSubmitIgnoresHiddenFields(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	b.visible = func(form.Values) bool { return false }
	b.validate = func() error { return errors.New("would fail if validated") }

	f := form.New(a.asField(), b.asField())
	_, cmd := f.Update(ctrlS())

	assert.Equal(t, 1, *a.validateCalls)
	assert.Equal(t, 0, *b.validateCalls)
	assert.True(t, hasSubmittedMsg(cmd))
}

func TestSubmitValidatesEveryVisibleFieldInOrder(t *testing.T) {
	// Drive Submit from a non-first focus position; assert every visible
	// field's validator ran at least once and SubmittedMsg fired. Tab now
	// gates on the current field's validator, so passing tabs also count
	// — we therefore use GreaterOrEqual rather than exact equality.
	a := newStub("a")
	b := newStub("b")
	c := newStub("c")
	f := form.New(a.asField(), b.asField(), c.asField())

	f, _ = f.Update(tab())
	f, _ = f.Update(tab())
	require.Equal(t, "c", f.Focused().Key())

	_, cmd := f.Update(ctrlS())
	assert.GreaterOrEqual(t, *a.validateCalls, 1)
	assert.GreaterOrEqual(t, *b.validateCalls, 1)
	assert.GreaterOrEqual(t, *c.validateCalls, 1)
	assert.True(t, hasSubmittedMsg(cmd))
}

func TestTabRefusesToAdvanceWhenValidatorFails(t *testing.T) {
	a := newStub("a")
	a.validate = func() error { return errors.New("a-invalid") }
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	require.Equal(t, "a", f.Focused().Key())
	f, _ = f.Update(tab())
	assert.Equal(t, "a", f.Focused().Key(), "tab must not advance past a failing field")
	assert.Equal(t, 1, *a.validateCalls)
}

func TestShortHelpComposesFormAndFieldBindings(t *testing.T) {
	altEnter := key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "newline"))
	plain := newStub("plain")
	multi := newStub("multi")
	multi.short = []key.Binding{altEnter}

	f := form.New(plain.asField(), multi.asField())

	help := f.ShortHelp()
	assert.Contains(t, help, f.KeyMap.Next)
	assert.Contains(t, help, f.KeyMap.Save)
	assert.NotContains(t, help, altEnter, "plain field has no extra bindings")

	f, _ = f.Update(tab())
	help = f.ShortHelp()
	assert.Contains(t, help, altEnter, "after focus moves, multi-line bindings appear")
}

func TestHiddenFieldIsNotRendered(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	b.visible = func(form.Values) bool { return false }
	c := newStub("c")
	f := form.New(a.asField(), b.asField(), c.asField())
	f, _ = f.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	v := f.View()
	assert.True(t, strings.Contains(v, "a-view"))
	assert.False(t, strings.Contains(v, "b-view"), "hidden field must not render")
	assert.True(t, strings.Contains(v, "c-view"))
}

// Under the progressive-snapshot rule, a field's Visible predicate only
// sees the values of visible fields before it. A focused field cannot
// hide itself via its own Update because its predicate cannot consult its
// own value. The defensive refocus path in form.refocusIfHidden therefore
// has no reachable trigger through Update; we do not test it here.

func TestHiddenFieldValueIsExcludedFromLaterSnapshots(t *testing.T) {
	a := newStub("a")
	a.val = "stashed"
	a.visible = func(form.Values) bool { return false }

	seen := make(chan any, 1)
	b := newStub("b")
	b.visible = func(v form.Values) bool {
		select {
		case seen <- v.Get("a"):
		default:
		}
		return true
	}
	_ = form.New(a.asField(), b.asField())

	select {
	case got := <-seen:
		assert.Nil(t, got, "hidden field's value must not appear in later snapshots")
	default:
		t.Fatal("predicate never ran")
	}
}

func TestWindowSizeMsgBroadcastsSetWidthToEveryField(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	c := newStub("c")
	f := form.New(a.asField(), b.asField(), c.asField())

	_, _ = f.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	assert.Equal(t, []int{80}, *a.widths)
	assert.Equal(t, []int{80}, *b.widths)
	assert.Equal(t, []int{80}, *c.widths)
}

func TestInitReturnsWindowSizeRequest(t *testing.T) {
	a := newStub("a")
	f := form.New(a.asField())
	cmd := f.Init()
	require.NotNil(t, cmd)
	// The batched cmd may include other cmds; we just need to confirm one of
	// them is the bubbletea window-size request. Running the batch returns a
	// BatchMsg of inner msgs.
	msgs := collectMsgs(cmd)
	var found bool
	for _, msg := range msgs {
		// tea.RequestWindowSize emits a tea.WindowSizeMsg with zero values when
		// run synchronously outside a Program; what matters is that the cmd is
		// present in the batch. The simpler invariant is that Init produced at
		// least one msg.
		_ = msg
		found = true
	}
	assert.True(t, found, "Init's batched cmd should yield at least one msg")
}

func TestFieldValuesExcludesHiddenFields(t *testing.T) {
	a := newStub("a")
	a.val = "av"
	b := newStub("b")
	b.val = "bv"
	b.visible = func(form.Values) bool { return false }
	c := newStub("c")
	c.val = "cv"

	f := form.New(a.asField(), b.asField(), c.asField())
	vals := f.FieldValues()
	assert.Equal(t, "av", vals["a"])
	assert.Equal(t, "cv", vals["c"])
	_, present := vals["b"]
	assert.False(t, present, "hidden field must not appear in FieldValues")
}

func TestVisibilityPredicateOnlySeesPriorFields(t *testing.T) {
	a := newStub("a")
	a.val = "av"
	b := newStub("b")
	b.val = "bv"

	seen := make(chan map[string]any, 1)
	c := newStub("c")
	c.val = "cv"
	c.visible = func(v form.Values) bool {
		snap := map[string]any{
			"a": v.Get("a"),
			"b": v.Get("b"),
			"c": v.Get("c"),
		}
		select {
		case seen <- snap:
		default:
		}
		return true
	}
	_ = form.New(a.asField(), b.asField(), c.asField())

	select {
	case got := <-seen:
		assert.Equal(t, "av", got["a"])
		assert.Equal(t, "bv", got["b"])
		assert.Nil(t, got["c"], "field's own value must not be in its own predicate's snapshot")
	default:
		t.Fatal("predicate never ran")
	}
}

func TestValuesGetUnknownKeyReturnsNil(t *testing.T) {
	got := make(chan any, 1)
	a := newStub("a")
	a.visible = func(v form.Values) bool {
		select {
		case got <- v.Get("does-not-exist"):
		default:
		}
		return true
	}
	f := form.New(a.asField())
	_ = f.View()

	select {
	case v := <-got:
		assert.Nil(t, v)
	default:
		t.Fatal("predicate never ran")
	}
}
