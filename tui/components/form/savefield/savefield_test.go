package savefield_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
)

func TestNewRequiresKey(t *testing.T) {
	assert.Panics(t, func() { savefield.New("") })
}

func TestDefaultLabel(t *testing.T) {
	m := savefield.New("save")
	assert.Contains(t, m.View(), "  Save  ")
}

func TestWithLabelOverrides(t *testing.T) {
	m := savefield.New("save", savefield.WithLabel("Update"))
	assert.Contains(t, m.View(), "  Update  ")
}

func TestValidateAlwaysPasses(t *testing.T) {
	m := savefield.New("save")
	_, err := m.Validate()
	assert.NoError(t, err)
}

func TestEnterDoesNotEmitFromField(t *testing.T) {
	// The savefield is a valueless placeholder: it no longer emits a submit
	// request of its own. The form owns Enter-to-submit at the last field.
	m := savefield.New("save")
	f, _ := m.Focus()
	_, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Nil(t, cmd, "savefield must not emit a submit request from its own Update")
}

func TestEnterIgnoredWhenUnfocused(t *testing.T) {
	m := savefield.New("save")
	var f form.Field = m
	_, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Nil(t, cmd, "unfocused savefield must not emit a submit request")
}

func TestEnterOnSavefieldSubmitsForm(t *testing.T) {
	// End-to-end: a form with one inputfield + savefield. Tab to the save
	// button and press Enter; the form's last-field rule submits directly
	// (no SubmitRequestMsg round-trip), emitting SubmittedMsg in the batch.
	title := inputfield.New("title", "Title", inputfield.WithValue("hi"))
	save := savefield.New("save")

	f := form.New(title, save)
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Equal(t, "save", f.Focused().Key())

	_, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.True(t, hasSubmittedMsg(cmd), "Enter on the terminal savefield submits the form")
}

func TestTabLeavesSavefieldWithoutSubmitting(t *testing.T) {
	title := inputfield.New("title", "Title", inputfield.WithValue("hi"))
	save := savefield.New("save")

	f := form.New(title, save)
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	require.Equal(t, "save", f.Focused().Key())

	// Tab on the last field is a no-op; shift+tab goes back. Either way
	// no submission fires.
	_, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.False(t, hasSubmittedMsg(cmd))

	_, cmd = f.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assert.False(t, hasSubmittedMsg(cmd))
}

func TestKeysDoesNotClaimEnter(t *testing.T) {
	// The placeholder claims no keys, so the form's last-field rule sees
	// Enter and submits. (A non-empty Keys() claiming Enter would route the
	// gesture to the field and defeat that rule.)
	m := savefield.New("save")
	assert.Empty(t, m.Keys(), "savefield is a valueless placeholder and claims no keys")
}

func TestVisibleDefaultTrue(t *testing.T) {
	m := savefield.New("save")
	assert.True(t, m.Visible(nil))
}

func TestWithVisiblePredicate(t *testing.T) {
	m := savefield.New("save", savefield.WithVisible(func(v form.Values) bool {
		return v.Get("ready") == true
	}))
	vals := stubValues{"ready": false}
	assert.False(t, m.Visible(vals))
	vals["ready"] = true
	assert.True(t, m.Visible(vals))
}

// hasSubmittedMsg walks a possibly-batched cmd and reports whether any
// emitted msg is form.SubmittedMsg.
func hasSubmittedMsg(cmd tea.Cmd) bool {
	for _, m := range collect(cmd) {
		if _, ok := m.(form.SubmittedMsg); ok {
			return true
		}
	}
	return false
}

func collect(cmd tea.Cmd) []tea.Msg {
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
			out = append(out, collect(c)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

type stubValues map[string]any

func (s stubValues) Get(k string) any { return s[k] }
