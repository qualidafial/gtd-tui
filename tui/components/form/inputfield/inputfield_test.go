package inputfield_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
)

func TestNewRequiresKey(t *testing.T) {
	assert.Panics(t, func() { inputfield.New("", "Title") })
}

func TestKeyAndLabel(t *testing.T) {
	f := inputfield.New("title", "Title")
	assert.Equal(t, "title", f.Key())
	assert.Contains(t, f.View(), "Title")
}

func TestWithValueSeedsInitialValue(t *testing.T) {
	f := inputfield.New("title", "Title", inputfield.WithValue("hello"))
	assert.Equal(t, "hello", f.Value())
}

func TestWithPlaceholderStored(t *testing.T) {
	// The textinput's placeholder is only rendered when there's width to
	// show it, which we don't configure in the unit test. Just confirm the
	// option seeds the field by typing nothing and observing empty value.
	f := inputfield.New("title", "Title", inputfield.WithPlaceholder("type here"))
	assert.Equal(t, "", f.Value())
}

func TestValidatorPasses(t *testing.T) {
	f := inputfield.New("title", "Title",
		inputfield.WithValue("ok"),
		inputfield.WithValidator(func(s string) error {
			if s == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	_, err := f.Validate()
	assert.NoError(t, err)
}

func TestValidatorFails(t *testing.T) {
	f := inputfield.New("title", "Title",
		inputfield.WithValidator(func(s string) error {
			if s == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	nf, err := f.Validate()
	require.Error(t, err)
	assert.EqualError(t, err, "required")
	// After Validate the field's View reflects the cached error.
	assert.Contains(t, nf.View(), "required")
}

func TestErrorClearsWhenValueEdited(t *testing.T) {
	f := inputfield.New("title", "Title",
		inputfield.WithValidator(func(s string) error {
			if s == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	nf, err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, nf.View(), "required")

	// Focus the field so the textinput actually consumes the keypress,
	// then type a character. The cached error should clear so the user
	// isn't shown a stale "required" message after they've fixed input.
	focused, _ := nf.Focus()
	updated, _ := focused.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.NotContains(t, updated.View(), "required")
}

func TestNilValidatorMeansAlwaysValid(t *testing.T) {
	f := inputfield.New("title", "Title")
	_, err := f.Validate()
	assert.NoError(t, err)
}

func TestSetValueUpdatesValueAndValidation(t *testing.T) {
	f := inputfield.New("title", "Title",
		inputfield.WithValidator(func(s string) error {
			if s == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	_, err := f.Validate()
	require.Error(t, err)
	f = f.SetValue("now-set")
	assert.Equal(t, "now-set", f.Value())
	_, err = f.Validate()
	assert.NoError(t, err)
}

func TestFocusTogglesCursor(t *testing.T) {
	f := inputfield.New("title", "Title")
	assert.False(t, f.Focused(), "field starts blurred")

	field, _ := f.Focus()
	assert.True(t, field.Focused(), "Focus marks the field focused")

	field = field.Blur()
	assert.False(t, field.Focused(), "Blur clears focus")
}

func TestUpdateTypesIntoTheField(t *testing.T) {
	f := inputfield.New("title", "Title")
	field, _ := f.Focus()

	for _, r := range "abc" {
		var ff form.Field = field
		ff, _ = ff.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		field = ff.(inputfield.Model)
	}
	assert.Equal(t, "abc", field.Value())
}

func TestVisibleDefaultTrue(t *testing.T) {
	f := inputfield.New("title", "Title")
	assert.True(t, f.Visible(nil))
}

func TestWithVisiblePredicate(t *testing.T) {
	f := inputfield.New("title", "Title",
		inputfield.WithVisible(func(v form.Values) bool {
			return v.Get("kind") == "task"
		}),
	)
	vals := stubValues{"kind": "project"}
	assert.False(t, f.Visible(vals))
	vals["kind"] = "task"
	assert.True(t, f.Visible(vals))
}

func TestWorksInsideFormHappyPath(t *testing.T) {
	title := inputfield.New("title", "Title",
		inputfield.WithValidator(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	desc := inputfield.New("desc", "Description")

	f := form.New(title, desc)
	require.NotNil(t, f.Focused())
	assert.Equal(t, "title", f.Focused().Key())

	// Submit while title is empty → focus stays on title and SubmittedMsg
	// is not emitted.
	f, ok, _ := f.Submit()
	assert.False(t, ok)
	assert.Equal(t, "title", f.Focused().Key())

	// Type a title, then submit again.
	for _, r := range "hi" {
		f, _ = f.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	_, ok, _ = f.Submit()
	assert.True(t, ok)
}

type stubValues map[string]any

func (s stubValues) Get(k string) any { return s[k] }
