package keymap

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chord builds a Chord with the given visibility and an optional Show
// subset (nil ⇒ all keys).
func chord(vis Vis, desc string, show []string, keys ...string) Chord {
	return Chord{
		Binding: key.NewBinding(key.WithKeys(keys...), key.WithHelp("", desc)),
		Show:    show,
		Vis:     vis,
	}
}

// labels flattens a resolved set to (key,desc) pairs in order for easy
// assertions.
type label struct {
	key, desc string
}

func labels(groups []Group) []label {
	var out []label
	for _, g := range groups {
		for _, c := range g {
			h := c.Help()
			out = append(out, label{h.Key, h.Desc})
		}
	}
	return out
}

func TestResolve_PartialShadowRelabels(t *testing.T) {
	field := Group{chord(Short, "select", nil, "down")}
	form := Group{chord(Short, "next", nil, "tab", "down")}

	got := Resolve(nil, field, form)

	assert.Equal(t, []label{
		{"↓", "select"},
		{"tab", "next"},
	}, labels(got))
}

func TestResolve_FullShadowDropsChord(t *testing.T) {
	inner := Group{chord(Short, "cancel", nil, "esc")}
	overlay := Group{chord(Short, "back", nil, "esc")}

	got := Resolve(nil, inner, overlay)

	assert.Equal(t, []label{{"esc", "cancel"}}, labels(got))
}

func TestResolve_HiddenAliasNeverSurfaced(t *testing.T) {
	// Keys{j,down}, Show{down}: j routes but never displays.
	field := Group{chord(Short, "move", []string{"down"}, "j", "down")}

	got := Resolve(nil, field)

	for _, l := range labels(got) {
		assert.NotContains(t, l.key, "j")
	}
	assert.Equal(t, []label{{"↓", "move"}}, labels(got))
}

func TestResolve_ClaimAccumulatesAcrossGroups(t *testing.T) {
	field := Group{chord(Short, "select", nil, "down")}
	form := Group{chord(Short, "next", nil, "tab", "down")}
	overlay := Group{chord(Short, "scroll", nil, "down")}

	got := Resolve(nil, field, form, overlay)

	// field keeps down; form's down removed (tab survives); overlay's
	// only key down removed → dropped.
	assert.Equal(t, []label{
		{"↓", "select"},
		{"tab", "next"},
	}, labels(got))
}

func TestResolve_DisabledChordClaimsAndShowsNothing(t *testing.T) {
	disabled := key.NewBinding(key.WithKeys("down"), key.WithHelp("", "select"))
	disabled.SetEnabled(false)
	field := Group{{Binding: disabled, Vis: Short}}
	form := Group{chord(Short, "next", nil, "tab", "down")}

	got := Resolve(nil, field, form)

	// Disabled field chord neither displays nor claims; form keeps down.
	assert.Equal(t, []label{{"tab/↓", "next"}}, labels(got))
}

func TestResolve_RouteOnlyClaimsButNeverDisplays(t *testing.T) {
	field := Group{chord(RouteOnly, "select", nil, "down")}
	form := Group{chord(Short, "next", nil, "tab", "down")}

	got := Resolve(nil, field, form)

	// RouteOnly field chord claims down (removed from form) but is itself
	// never displayed in either help bar.
	short := ShortHelp(got)
	require.Len(t, short, 1)
	assert.Equal(t, "next", short[0].Help().Desc)
	assert.Equal(t, "tab", short[0].Help().Key)
}

func TestResolve_NonMutation(t *testing.T) {
	field := Group{chord(Short, "select", nil, "down")}
	form := Group{chord(Short, "next", nil, "tab", "down")}

	before := labels([]Group{field, form})
	_ = Resolve(nil, field, form)
	_ = Resolve(nil, field, form)
	after := labels([]Group{field, form})

	assert.Equal(t, before, after, "inputs must be unchanged across calls")
}

func TestResolve_GroupOrderPreserved(t *testing.T) {
	g1 := Group{chord(Short, "a", nil, "a")}
	g2 := Group{chord(Short, "b", nil, "b")}
	g3 := Group{chord(Short, "c", nil, "c")}

	got := Resolve(nil, g1, g2, g3)

	require.Len(t, got, 3)
	assert.Equal(t, "a", got[0][0].Help().Desc)
	assert.Equal(t, "b", got[1][0].Help().Desc)
	assert.Equal(t, "c", got[2][0].Help().Desc)
}

func TestProjections_ShortFiltersAndFlattens(t *testing.T) {
	g := Group{
		chord(Short, "short", nil, "a"),
		chord(Full, "full", nil, "b"),
		chord(RouteOnly, "route", nil, "c"),
	}
	resolved := Resolve(nil, g)

	short := ShortHelp(resolved)
	require.Len(t, short, 1)
	assert.Equal(t, "short", short[0].Help().Desc)
}

func TestProjections_FullKeepsGroupsAndShortPlusFull(t *testing.T) {
	g1 := Group{chord(Short, "short", nil, "a"), chord(Full, "full", nil, "b")}
	g2 := Group{chord(RouteOnly, "route", nil, "c")}
	resolved := Resolve(nil, g1, g2)

	full := FullHelp(resolved)
	// g2 has only a RouteOnly chord → no row.
	require.Len(t, full, 1)
	require.Len(t, full[0], 2)
	assert.Equal(t, "short", full[0][0].Help().Desc)
	assert.Equal(t, "full", full[0][1].Help().Desc)
}

func TestProjections_RouteOnlyExcludedFromBoth(t *testing.T) {
	g := Group{chord(RouteOnly, "route", nil, "a")}
	resolved := Resolve(nil, g)

	assert.Empty(t, ShortHelp(resolved))
	assert.Empty(t, FullHelp(resolved))
}

// stubMap adapts groups to the Map interface for Handles tests.
type stubMap struct{ groups []Group }

func (s stubMap) Chords() []Group { return s.groups }

func TestHandles_MatchByCompleteKeysInclHiddenAlias(t *testing.T) {
	child := stubMap{groups: []Group{{chord(Short, "move", []string{"down"}, "j", "down")}}}

	// Hidden alias j still routes.
	assert.True(t, Handles(child, tea.KeyPressMsg{Code: 'j', Text: "j"}))
	assert.True(t, Handles(child, tea.KeyPressMsg{Code: tea.KeyDown}))
	assert.False(t, Handles(child, tea.KeyPressMsg{Code: tea.KeyUp}))
}

func TestHandles_DisabledReturnsFalse(t *testing.T) {
	b := key.NewBinding(key.WithKeys("down"))
	b.SetEnabled(false)
	child := stubMap{groups: []Group{{{Binding: b, Vis: Short}}}}

	assert.False(t, Handles(child, tea.KeyPressMsg{Code: tea.KeyDown}))
}

func TestHandles_DepthViaAggregatedChords(t *testing.T) {
	// Simulate an aggregated subtree: field group first, then parent group.
	child := stubMap{groups: []Group{
		{chord(Short, "help", nil, "?")},
		{chord(Short, "next", nil, "tab")},
	}}

	assert.True(t, Handles(child, tea.KeyPressMsg{Code: '?', Text: "?"}))
	assert.True(t, Handles(child, tea.KeyPressMsg{Code: tea.KeyTab}))
}

func TestDefaultRender_Glyphs(t *testing.T) {
	out := DefaultRender([]string{"shift+tab", "up"})
	assert.Contains(t, out, "⇧tab")
	assert.Contains(t, out, "↑")
}

func TestDefaultRender_Fallback(t *testing.T) {
	out := DefaultRender([]string{"ctrl+x"})
	assert.Contains(t, out, "ctrl+x")
}
