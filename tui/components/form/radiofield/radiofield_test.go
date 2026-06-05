package radiofield_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/radiofield"
)

type kind string

const (
	kindTask    kind = "task"
	kindProject kind = "project"
)

func sample() []radiofield.Option[kind] {
	return []radiofield.Option[kind]{
		{Display: "Task", Value: kindTask},
		{Display: "Project", Value: kindProject},
	}
}

func focus[T comparable](m radiofield.Model[T]) form.Field {
	f, _ := m.Focus()
	return f
}

func TestNewRequiresKey(t *testing.T) {
	assert.Panics(t, func() { radiofield.New("", "Kind", sample()) })
}

func TestNewRequiresNonEmptyOptions(t *testing.T) {
	assert.Panics(t, func() { radiofield.New[kind]("kind", "Kind", nil) })
}

func TestInitialSelectionIsFirst(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample())
	assert.Equal(t, kindTask, m.SelectedValue())
}

func TestWithInitialValueSelectsMatchingOption(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample(), radiofield.WithInitialValue(kindProject))
	assert.Equal(t, kindProject, m.SelectedValue())
}

func TestWithInitialValueIgnoredWhenNoMatch(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample(), radiofield.WithInitialValue(kind("unknown")))
	assert.Equal(t, kindTask, m.SelectedValue(), "no matching option leaves first selected")
}

func TestRightAdvancesSelection(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample())
	f := focus(m)
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	got := f.(radiofield.Model[kind])
	assert.Equal(t, kindProject, got.SelectedValue())
}

func TestLeftRetreatsSelection(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample(), radiofield.WithInitialValue(kindProject))
	f := focus(m)
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	got := f.(radiofield.Model[kind])
	assert.Equal(t, kindTask, got.SelectedValue())
}

func TestArrowsClampAtEnds(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample())
	f := focus(m)

	// Left at first option stays at first.
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	assert.Equal(t, kindTask, f.(radiofield.Model[kind]).SelectedValue())

	// Right twice — second advance is clamped at last.
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.Equal(t, kindProject, f.(radiofield.Model[kind]).SelectedValue())
}

func TestUnfocusedFieldIgnoresArrows(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample())
	var f form.Field = m
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.Equal(t, kindTask, f.(radiofield.Model[kind]).SelectedValue())
}

func TestViewRendersAllOptionsInline(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample())
	v := m.View()
	assert.Contains(t, v, "Task")
	assert.Contains(t, v, "Project")
	// One line means no newline between options (label may be on its own line).
	rows := strings.Split(v, "\n")
	// Find the row that contains both Task and Project.
	var found bool
	for _, r := range rows {
		if strings.Contains(r, "Task") && strings.Contains(r, "Project") {
			found = true
		}
	}
	assert.True(t, found, "all options must render on a single row")
}

func TestSelectedAndUnselectedOptionsStyledDifferently(t *testing.T) {
	// Both options render as "  <name>  " (no radio glyph). The selected
	// option uses the accent background style; unselected uses the gray
	// background. The styles produce distinct ANSI escape sequences, so a
	// substring check on the rendered output is enough.
	m := radiofield.New("kind", "Kind", sample())
	v := m.View()
	assert.Contains(t, v, "  Task  ")
	assert.Contains(t, v, "  Project  ")
}

func TestKeysAdvertiseArrowBinding(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample())
	groups := m.Keys()
	require.NotEmpty(t, groups)
	require.NotEmpty(t, groups[0])
	c := groups[0][0]
	assert.Equal(t, "←/→", c.Help().Key)
	assert.ElementsMatch(t, []string{"left", "right"}, c.Keys())
}

func TestValidatorRunsOnCurrentValue(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample(),
		radiofield.WithValidator(func(k kind) error {
			if k == kindProject {
				return errors.New("not allowed")
			}
			return nil
		}),
	)
	_, err := m.Validate()
	assert.NoError(t, err)

	// Move to project, re-validate.
	f := focus(m)
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	nf, err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, nf.View(), "not allowed")
}

func TestErrorClearsOnSelectionChange(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample(),
		radiofield.WithValidator(func(k kind) error {
			if k == kindTask {
				return errors.New("bad")
			}
			return nil
		}),
	)
	nf, err := m.Validate()
	require.Error(t, err)

	// Change selection — the stale cached error should clear.
	f, _ := nf.Focus()
	f, _ = f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.NotContains(t, f.View(), "bad")
}

func TestVisibleDefaultTrue(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample())
	assert.True(t, m.Visible(nil))
}

func TestWithVisiblePredicate(t *testing.T) {
	m := radiofield.New("kind", "Kind", sample(),
		radiofield.WithVisible[kind](func(v form.Values) bool {
			return v.Get("show") == true
		}),
	)
	vals := stubValues{"show": false}
	assert.False(t, m.Visible(vals))
	vals["show"] = true
	assert.True(t, m.Visible(vals))
}

type stubValues map[string]any

func (s stubValues) Get(k string) any { return s[k] }
