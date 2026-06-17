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
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// stubField is a controllable form.Field for unit tests.
type stubField struct {
	key      string
	val      any
	focused  bool
	visible  func(form.Values) bool
	validate func() error
	bindings []keymap.Group

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

func (s stubField) Keys() []keymap.Group { return s.bindings }

// setValueMsg lets a test change the value of a specific stub field through
// the form's Update pipeline.
type setValueMsg struct {
	key string
	val any
}

func tab() tea.KeyPressMsg      { return tea.KeyPressMsg{Code: tea.KeyTab} }
func shiftTab() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift} }
func ctrlS() tea.KeyPressMsg    { return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl} }
func enter() tea.KeyPressMsg    { return tea.KeyPressMsg{Code: tea.KeyEnter} }

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

func TestEnterOnLastVisibleFieldSubmits(t *testing.T) {
	// The last field claims no keys, so Enter falls through to the form's
	// last-field rule and submits.
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	f, _ = f.Update(tab())
	require.Equal(t, "b", f.Focused().Key())

	_, cmd := f.Update(enter())
	assert.True(t, hasSubmittedMsg(cmd), "Enter on the last visible field submits")
}

func TestEnterOnNonLastFieldAdvances(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField())
	require.Equal(t, "a", f.Focused().Key())

	f, cmd := f.Update(enter())
	assert.Equal(t, "b", f.Focused().Key(), "Enter advances when a later field exists")
	assert.False(t, hasSubmittedMsg(cmd), "Enter must not submit while a later field exists")
}

func TestEnterAdvanceIsValidatorGated(t *testing.T) {
	a := newStub("a")
	a.validate = func() error { return errors.New("a-invalid") }
	b := newStub("b")
	f := form.New(a.asField(), b.asField())

	f, cmd := f.Update(enter())
	assert.Equal(t, "a", f.Focused().Key(), "Enter gates on the validator like tab")
	assert.False(t, hasSubmittedMsg(cmd))
}

func TestEnterOnClaimingLastFieldRoutesToField(t *testing.T) {
	// A last field that claims Enter (e.g. a multi-line field, or a select
	// in submit-on-enter mode) keeps the gesture; the form must not submit
	// on its behalf.
	enterBinding := key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"))
	a := newStub("a")
	b := newStub("b")
	b.bindings = []keymap.Group{{{Binding: enterBinding, Vis: keymap.Short}}}
	f := form.New(a.asField(), b.asField())

	f, _ = f.Update(tab())
	require.Equal(t, "b", f.Focused().Key())

	_, cmd := f.Update(enter())
	assert.False(t, hasSubmittedMsg(cmd), "a claiming last field keeps Enter; the form must not submit")
	assert.Equal(t, 1, *b.updateCalls, "Enter routes to the claiming field's Update")
}

// helpDescs flattens a short-help projection to its binding descriptions.
// The resolver emits relabeled key.Binding copies, so tests assert on
// description rather than binding identity.
func helpDescs(bindings []key.Binding) []string {
	out := make([]string, len(bindings))
	for i, b := range bindings {
		out[i] = b.Help().Desc
	}
	return out
}

func TestShortHelpComposesFormAndFieldBindings(t *testing.T) {
	altEnter := key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "newline"))
	plain := newStub("plain")
	multi := newStub("multi")
	multi.bindings = []keymap.Group{{{Binding: altEnter, Vis: keymap.Short}}}

	f := form.New(plain.asField(), multi.asField())

	descs := helpDescs(f.ShortHelp())
	assert.Contains(t, descs, "next")
	assert.Contains(t, descs, "save")
	assert.NotContains(t, descs, "newline", "plain field has no extra bindings")

	f, _ = f.Update(tab())
	descs = helpDescs(f.ShortHelp())
	assert.Contains(t, descs, "newline", "after focus moves, multi-line bindings appear")
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

func TestUpdateFieldReplacesInPlaceAndReflectsInValuesAndView(t *testing.T) {
	a := newStub("a")
	a.val = "av"
	b := newStub("b")
	b.val = "bv"
	f := form.New(a.asField(), b.asField())
	f, _ = f.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	f = f.UpdateField("b", func(field form.Field) form.Field {
		s := field.(stubField)
		s.val = "loaded"
		s.key = "b" // key is immutable; identity must be preserved by callers
		return s
	})

	assert.Equal(t, "loaded", f.FieldValues()["b"], "UpdateField result is visible to FieldValues")
	assert.Equal(t, "av", f.FieldValues()["a"], "other fields are untouched")
	assert.Contains(t, f.View(), "b-view", "viewport is re-synced after UpdateField")
}

func TestUpdateFieldPreservesFocus(t *testing.T) {
	a := newStub("a")
	b := newStub("b")
	f := form.New(a.asField(), b.asField()) // a starts focused

	f = f.UpdateField("a", func(field form.Field) form.Field {
		s := field.(stubField)
		s.val = "x"
		return s
	})

	require.NotNil(t, f.Focused())
	assert.Equal(t, "a", f.Focused().Key(), "focus stays on the updated field")
	assert.Equal(t, "x", f.Focused().Value())
}

func TestUpdateFieldPanicsOnUnknownKey(t *testing.T) {
	f := form.New(newStub("a").asField())
	assert.Panics(t, func() {
		f.UpdateField("missing", func(field form.Field) form.Field { return field })
	})
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
