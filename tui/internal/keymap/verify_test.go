package keymap_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/radiofield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/selectfield"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

func descKeys(bindings []key.Binding) map[string]string {
	out := map[string]string{}
	for _, b := range bindings {
		out[b.Help().Desc] = b.Help().Key
	}
	return out
}

// 7.2: a focused selectfield shows ↑/↓ once (under "choose"), and the
// form's next-field binding is relabeled without ↓.
func TestVerify_SelectfieldArrowsDedup(t *testing.T) {
	sf := selectfield.New("kind", "Kind", []selectfield.Option[string]{
		{Display: "a", Value: "a"},
		{Display: "b", Value: "b"},
	})
	f := form.New(sf, savefield.New("save"))

	short := descKeys(f.ShortHelp())
	require.Contains(t, short, "choose")
	assert.Equal(t, "↑/↓", short["choose"], "field owns the arrows")
	// form's next survives but without ↓ (tab remains).
	require.Contains(t, short, "next")
	assert.Equal(t, "tab", short["next"], "next relabeled without ↓")
}

// 7.2: radiofield owns ←/→.
func TestVerify_RadiofieldArrows(t *testing.T) {
	rf := radiofield.New("k", "K", []radiofield.Option[string]{
		{Display: "x", Value: "x"},
		{Display: "y", Value: "y"},
	})
	f := form.New(rf, savefield.New("save"))
	short := descKeys(f.ShortHelp())
	assert.Equal(t, "←/→", short["choose"])
}

// 7.3: a key bound by both field and form routes to the field. The
// savefield claims enter; the form must forward enter (Handles true) so it
// is not treated as next-field.
func TestVerify_SavefieldClaimsEnter(t *testing.T) {
	sf := savefield.New("save")
	var fld form.Field = sf
	fld, _ = fld.Focus()
	assert.True(t, keymap.Handles(fld, tea.KeyPressMsg{Code: tea.KeyEnter}),
		"savefield must claim enter so the form forwards it")

	// A plain inputfield does NOT claim enter — the form may advance.
	var in form.Field = inputfield.New("title", "Title")
	in, _ = in.Focus()
	assert.False(t, keymap.Handles(in, tea.KeyPressMsg{Code: tea.KeyEnter}),
		"inputfield does not claim enter")
}

// 7.3: deep delegation — a form aggregates the focused field's bindings, so
// Handles over the form reports the field's claimed keys.
func TestVerify_DeepDelegationThroughForm(t *testing.T) {
	sf := selectfield.New("kind", "Kind", []selectfield.Option[string]{
		{Display: "a", Value: "a"},
	})
	f := form.New(sf)
	assert.True(t, keymap.Handles(f, tea.KeyPressMsg{Code: tea.KeyDown}),
		"form's aggregated bindings carry the field's down claim")
}
