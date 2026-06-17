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

func TestKeysAdvertiseFilterAndClaimArrows(t *testing.T) {
	// Filtering is inherited from list.Model; we don't re-test bubbles'
	// fuzzy-filter behavior here. The contract for selectfield is that the
	// `/`-to-filter binding shows up in the form's help footer and that
	// up/down are claimed so they route to the list, not form navigation.
	opts := []selectfield.Option[string]{
		{Display: "apple", Value: "apple"},
		{Display: "banana", Value: "banana"},
	}
	m := selectfield.New("fruit", "Fruit", opts)

	var foundFilter, claimsUp, claimsDown bool
	for _, g := range m.Keys() {
		for _, c := range g {
			for _, k := range c.Keys() {
				switch k {
				case "/":
					foundFilter = true
				case "up":
					claimsUp = true
				case "down":
					claimsDown = true
				}
			}
		}
	}
	assert.True(t, foundFilter, "selectfield Keys should expose the filter binding")
	assert.True(t, claimsUp && claimsDown, "selectfield Keys should claim up/down")
}

func claimsEnter[T comparable](m selectfield.Model[T]) bool {
	for _, g := range m.Keys() {
		for _, c := range g {
			for _, k := range c.Keys() {
				if k == "enter" {
					return true
				}
			}
		}
	}
	return false
}

func TestKeysClaimEnterOnlyWhileFiltering(t *testing.T) {
	// Not filtering: Enter is unclaimed so the form's last-field rule can
	// submit a terminal selectfield. While filtering: Enter is claimed so it
	// routes to the list and accepts the filter instead of submitting.
	opts := []selectfield.Option[string]{
		{Display: "apple", Value: "apple"},
		{Display: "banana", Value: "banana"},
	}
	m := selectfield.New("fruit", "Fruit", opts)
	m = m.SetWidth(40).(selectfield.Model[string])

	assert.False(t, claimsEnter(m), "a non-filtering selectfield must not claim enter")

	f := focus(m)
	f, _ = f.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	assert.True(t, claimsEnter(f.(selectfield.Model[string])),
		"a filtering selectfield claims enter to accept the filter")
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

func TestSetOptionsReplacesOptions(t *testing.T) {
	m := selectfield.New[kind]("kind", "Kind", nil)
	assert.Equal(t, kind(""), m.SelectedValue(), "no options yields the zero value")

	m = m.SetOptions(sample())
	m = m.SetWidth(40).(selectfield.Model[kind])
	assert.Equal(t, kindTask, m.SelectedValue(), "first option of the new set is selected")

	f := focus(m)
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got := f.(selectfield.Model[kind])
	assert.Equal(t, kindProject, got.SelectedValue(), "dynamic options are navigable")
}

func TestSetOptionsReprependsNone(t *testing.T) {
	type id int64
	m := selectfield.New[*id]("p", "Project", nil, selectfield.WithNone[*id]("(none)"))
	m = m.SetWidth(40).(selectfield.Model[*id])
	assert.Nil(t, m.SelectedValue())
	assert.Contains(t, m.View(), "(none)")

	one := new(id(1))
	opts := []selectfield.Option[*id]{
		{Display: "alpha", Value: one},
		{Display: "beta", Value: new(id(2))},
	}
	m = m.SetOptions(opts)
	m = m.SetWidth(40).(selectfield.Model[*id])
	assert.Nil(t, m.SelectedValue(), "the re-prepended none option starts selected")
	assert.Contains(t, m.View(), "(none)")

	// Navigating past the none entry reaches the first real option.
	f := focus(m)
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got := f.(selectfield.Model[*id])
	assert.Equal(t, one, got.SelectedValue())
}

func TestSetOptionsReappliesInitialSelection(t *testing.T) {
	m := selectfield.New("kind", "Kind", nil,
		selectfield.WithInitialValue(kindProject),
	)
	m = m.SetOptions(sample())
	assert.Equal(t, kindProject, m.SelectedValue(),
		"WithInitialValue is re-applied against the new options")
}

func TestWithHideWhenEmpty(t *testing.T) {
	m := selectfield.New[kind]("kind", "Kind", nil, selectfield.WithHideWhenEmpty[kind]())
	assert.False(t, m.Visible(nil), "hidden while it has no real options")

	m = m.SetOptions(sample())
	assert.True(t, m.Visible(nil), "visible once options load")
}

func TestWithHideWhenEmptyIgnoresNoneEntry(t *testing.T) {
	m := selectfield.New[kind]("kind", "Kind", nil,
		selectfield.WithNone[kind]("(none)"),
		selectfield.WithHideWhenEmpty[kind](),
	)
	assert.False(t, m.Visible(nil), "a WithNone entry alone does not count as a real option")
}

func TestWithHideWhenEmptyComposesWithVisible(t *testing.T) {
	m := selectfield.New[kind]("kind", "Kind", sample(),
		selectfield.WithHideWhenEmpty[kind](),
		selectfield.WithVisible[kind](func(v form.Values) bool { return v.Get("on") == true }),
	)
	assert.False(t, m.Visible(stubValues{"on": false}), "predicate still gates a populated field")
	assert.True(t, m.Visible(stubValues{"on": true}))
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
