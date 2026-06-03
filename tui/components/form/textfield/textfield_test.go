package textfield_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/textfield"
)

func focusedField(t *testing.T, m textfield.Model) form.Field {
	t.Helper()
	f, _ := m.Focus()
	return f
}

func TestNewRequiresKey(t *testing.T) {
	assert.Panics(t, func() { textfield.New("", "Body") })
}

func TestKeyAndLabel(t *testing.T) {
	m := textfield.New("body", "Body")
	assert.Equal(t, "body", m.Key())
	assert.Contains(t, m.View(), "Body")
}

func TestWithValueSeedsInitialValue(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithValue("seed"))
	assert.Equal(t, "seed", m.Value())
}

func TestAltEnterInsertsNewline(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithValue("hello"))
	f := focusedField(t, m)

	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})

	assert.Contains(t, f.Value().(string), "\n", "alt+enter must insert a newline")
}

func TestCtrlJInsertsNewline(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithValue("hello"))
	f := focusedField(t, m)

	f, _ = f.Update(tea.KeyPressMsg{Code: 'j', Mod: tea.ModCtrl})

	assert.Contains(t, f.Value().(string), "\n", "ctrl+j must insert a newline")
}

func TestPlainEnterDoesNotInsertNewline(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithValue("hello"))
	f := focusedField(t, m)

	f, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, "hello", f.Value(), "plain enter must not modify value")
	assert.Nil(t, cmd, "plain enter must not produce a cmd")
	assert.NotContains(t, f.Value().(string), "\n")
}

func TestTabIsNotConsumed(t *testing.T) {
	// The form expects tab to bubble up to it for navigation. The
	// textarea must not eat the keypress.
	tf := textfield.New("body", "Body")
	other := textfield.New("other", "Other")

	f := form.New(tf, other)
	require.Equal(t, "body", f.Focused().Key())

	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, "other", f.Focused().Key(), "tab inside textfield should bubble up to form navigation")
}

func TestChordsAdvertiseNewlineBinding(t *testing.T) {
	m := textfield.New("body", "Body")
	groups := m.Chords()
	require.NotEmpty(t, groups)

	var found bool
	for _, g := range groups {
		for _, c := range g {
			for _, k := range c.Keys() {
				if k == "alt+enter" || k == "ctrl+j" {
					found = true
				}
			}
		}
	}
	assert.True(t, found, "textfield Chords must advertise the newline binding")
}

func TestValidatorPasses(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithValue("ok"),
		textfield.WithValidator(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	_, err := m.Validate()
	assert.NoError(t, err)
}

func TestValidatorFailsCachesErrorInView(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	}))
	nf, err := m.Validate()
	require.Error(t, err)
	assert.Contains(t, nf.View(), "required")
}

func TestErrorClearsWhenValueEdited(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithValidator(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	}))
	nf, err := m.Validate()
	require.Error(t, err)
	assert.Contains(t, nf.View(), "required")

	focused, _ := nf.Focus()
	updated, _ := focused.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.NotContains(t, updated.View(), "required")
}

func TestVisibleDefaultTrue(t *testing.T) {
	m := textfield.New("body", "Body")
	assert.True(t, m.Visible(nil))
}

func TestWithVisiblePredicate(t *testing.T) {
	m := textfield.New("body", "Body", textfield.WithVisible(func(v form.Values) bool {
		return v.Get("kind") == "task"
	}))
	vals := stubValues{"kind": "project"}
	assert.False(t, m.Visible(vals))
	vals["kind"] = "task"
	assert.True(t, m.Visible(vals))
}

type stubValues map[string]any

func (s stubValues) Get(k string) any { return s[k] }
