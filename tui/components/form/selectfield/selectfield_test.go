package selectfield_test

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/selectfield"
)

type kind string

const (
	kindTask    kind = "task"
	kindProject kind = "project"
)

func sample() []selectfield.Option[kind] {
	return []selectfield.Option[kind]{
		{Display: "Task", Value: kindTask},
		{Display: "Project", Value: kindProject},
	}
}

func focus[T comparable](m selectfield.Model[T]) form.Field {
	f, _ := m.Focus()
	return f
}

func TestNewRequiresKey(t *testing.T) {
	assert.Panics(t, func() { selectfield.New[kind]("", "Kind", sample()) })
}

func TestKeyAndLabel(t *testing.T) {
	m := selectfield.New("kind", "Kind", sample())
	m = m.SetWidth(40).(selectfield.Model[kind])
	assert.Equal(t, "kind", m.Key())
	assert.Contains(t, m.View(), "Kind")
}

func TestInitialSelection(t *testing.T) {
	m := selectfield.New("kind", "Kind", sample())
	assert.Equal(t, kindTask, m.SelectedValue(), "first option starts selected")
	assert.Equal(t, any(kindTask), m.Value())
}

func TestNavigateDownChangesSelection(t *testing.T) {
	m := selectfield.New("kind", "Kind", sample())
	m = m.SetWidth(40).(selectfield.Model[kind])
	f := focus(m)

	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got, _ := f.(selectfield.Model[kind])
	assert.Equal(t, kindProject, got.SelectedValue())
}

func TestNavigationDoesNothingWhenUnfocused(t *testing.T) {
	m := selectfield.New("kind", "Kind", sample())
	var f form.Field = m
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got := f.(selectfield.Model[kind])
	assert.Equal(t, kindTask, got.SelectedValue(), "an unfocused selectfield must not consume keys")
}

func TestShortHelpAdvertisesFilterBinding(t *testing.T) {
	// Filtering is inherited from list.Model; we don't re-test bubbles'
	// fuzzy-filter behavior here. The contract for selectfield is that the
	// `/`-to-filter binding shows up in the form's help footer when this
	// field is focused, which it does via list.ShortHelp.
	opts := []selectfield.Option[string]{
		{Display: "apple", Value: "apple"},
		{Display: "banana", Value: "banana"},
	}
	m := selectfield.New("fruit", "Fruit", opts)

	var found bool
	for _, b := range m.ShortHelp() {
		for _, k := range b.Keys() {
			if k == "/" {
				found = true
			}
		}
	}
	assert.True(t, found, "selectfield ShortHelp should expose the filter binding")
}

func TestWithNonePrependsZeroValueOption(t *testing.T) {
	type id int64
	opts := []selectfield.Option[*id]{
		{Display: "alpha", Value: new(id(1))},
		{Display: "beta", Value: new(id(2))},
	}
	m := selectfield.New("p", "Project", opts, selectfield.WithNone[*id]("(none)"))
	m = m.SetWidth(40).(selectfield.Model[*id])
	assert.Nil(t, m.SelectedValue(), "with WithNone, the none option starts selected")
	assert.Contains(t, m.View(), "(none)")
}

func TestValidatorPassesOnValidSelection(t *testing.T) {
	m := selectfield.New("kind", "Kind", sample(),
		selectfield.WithValidator(func(k kind) error {
			if k == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	_, err := m.Validate()
	assert.NoError(t, err)
}

func TestValidatorFailsCachesErrorInView(t *testing.T) {
	m := selectfield.New[kind]("kind", "Kind", nil,
		selectfield.WithValidator(func(k kind) error {
			if k == "" {
				return errors.New("required")
			}
			return nil
		}),
	)
	nf, err := m.Validate()
	require.Error(t, err)
	assert.Contains(t, nf.View(), "required")
}

func TestVisibleDefaultTrue(t *testing.T) {
	m := selectfield.New("k", "K", sample())
	assert.True(t, m.Visible(nil))
}

func TestWithVisiblePredicate(t *testing.T) {
	m := selectfield.New("k", "K", sample(),
		selectfield.WithVisible[kind](func(v form.Values) bool {
			return v.Get("on") == true
		}),
	)
	vals := stubValues{"on": false}
	assert.False(t, m.Visible(vals))
	vals["on"] = true
	assert.True(t, m.Visible(vals))
}

type stubValues map[string]any

func (s stubValues) Get(k string) any { return s[k] }
