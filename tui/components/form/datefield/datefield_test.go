package datefield_test

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/datefield"
)

func TestNewRequiresKey(t *testing.T) {
	assert.Panics(t, func() { datefield.New("", "When") })
}

func TestKeyAndLabel(t *testing.T) {
	m := datefield.New("when", "When")
	assert.Equal(t, "when", m.Key())
	assert.Contains(t, m.View(), "When")
}

func TestEmptyValueParsesToNil(t *testing.T) {
	m := datefield.New("when", "When")
	assert.Nil(t, m.SelectedTime())
	assert.Nil(t, m.Value())
}

func TestWithValueSeedsInitialValue(t *testing.T) {
	t0 := time.Date(2026, 6, 2, 15, 0, 0, 0, time.Local)
	m := datefield.New("when", "When", datefield.WithValue(&t0))
	got := m.SelectedTime()
	require.NotNil(t, got)
	assert.True(t, got.Equal(t0))
}

func TestAbsoluteDateParses(t *testing.T) {
	m := datefield.New("when", "When")
	f := focusedWithText(t, m, "2026-06-02 15:00")
	got := f.(datefield.Model).SelectedTime()
	require.NotNil(t, got)
	want := time.Date(2026, 6, 2, 15, 0, 0, 0, time.Local)
	assert.True(t, got.Equal(want), "expected %v, got %v", want, got)
}

func TestNaturalLanguageDateParses(t *testing.T) {
	m := datefield.New("when", "When")
	f := focusedWithText(t, m, "tomorrow")
	got := f.(datefield.Model).SelectedTime()
	require.NotNil(t, got, "natural-language phrases should parse to a non-nil time")
}

func TestBlurSnapsToAbsolute(t *testing.T) {
	m := datefield.New("when", "When")
	f := focusedWithText(t, m, "tomorrow")
	blurred := f.Blur()
	v := blurred.View()
	// After Blur the displayed text should be the absolute YYYY-MM-DD
	// (with optional time) form, not the literal "tomorrow".
	assert.NotContains(t, v, "tomorrow", "Blur must snap input to absolute representation")
}

func TestInvalidInputProducesError(t *testing.T) {
	m := datefield.New("when", "When")
	f := focusedWithText(t, m, "not a date at all")
	nf, err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, nf.View(), "invalid")
}

func TestValidatorRunsOnParsedTime(t *testing.T) {
	notInPast := func(t *time.Time) error {
		if t != nil && t.Before(time.Now().Add(-24*time.Hour)) {
			return errors.New("must not be in the past")
		}
		return nil
	}
	m := datefield.New("when", "When", datefield.WithValidator(notInPast))
	f := focusedWithText(t, m, "2020-01-01")
	nf, err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, nf.View(), "must not be in the past")
}

func TestErrorClearsWhenValueEdited(t *testing.T) {
	m := datefield.New("when", "When")
	f := focusedWithText(t, m, "garbage")
	nf, err := f.Validate()
	require.Error(t, err)

	// Edit the value — error should clear on the next View.
	updated, _ := nf.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	view := updated.View()
	assert.NotContains(t, view, "invalid", "stale parse error must clear after edit")
}

func TestVisibleDefaultTrue(t *testing.T) {
	m := datefield.New("when", "When")
	assert.True(t, m.Visible(nil))
}

func TestWithVisiblePredicate(t *testing.T) {
	m := datefield.New("when", "When",
		datefield.WithVisible(func(v form.Values) bool {
			return v.Get("kind") == "task"
		}),
	)
	vals := stubValues{"kind": "project"}
	assert.False(t, m.Visible(vals))
	vals["kind"] = "task"
	assert.True(t, m.Visible(vals))
}

// focusedWithText returns f focused and with `text` typed into it, by
// dispatching each rune as a key press through Update.
func focusedWithText(t *testing.T, m datefield.Model, text string) form.Field {
	t.Helper()
	var f form.Field
	f, _ = m.Focus()
	for _, r := range text {
		f, _ = f.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return f
}

type stubValues map[string]any

func (s stubValues) Get(k string) any { return s[k] }
