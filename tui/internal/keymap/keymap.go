// Package keymap is the stack-wide source of truth for keybinding
// ownership across priority-ordered layers (field → form → overlay →
// app). A gesture routes to its highest-priority owner, and help is a
// priority-merged, relabeled projection of the same per-layer binding
// data, so routing and help cannot drift.
//
// A [Binding] wraps a live charm.land/bubbles/v2/key.Binding (its triggers,
// description, and enabled state are read live, never copied) and adds two
// orthogonal display controls: [Binding.Show] (which of the binding's keys are
// named in help) and [Binding.Vis] (in which help bars the binding appears). A
// [Group] is both the unit of conflict resolution and one full-help
// column. A [Map] returns its [Group]s already aggregated — a composite
// concatenates its focused child's Keys() ahead of its own — so the
// returned slice is the whole active subtree flattened, highest-priority
// first, and a single [Handles] or [Resolve] call works at any depth with
// no separate stack walker.
package keymap

import "charm.land/bubbles/v2/key"

// Vis controls in which help bars a [Binding] appears. Values are declared
// in increasing order of visibility; the zero value is [RouteOnly] so a
// binding must opt in to being displayed. Vis affects display only — it
// never affects routing or claiming.
type Vis uint8

const (
	// RouteOnly bindings route and claim (subtract a lower layer's matching
	// key from help) but never appear in any help bar. This is the zero
	// value: a plain Binding{Binding: b} routes and claims but is shown
	// nowhere until its Vis is set.
	RouteOnly Vis = iota
	// Short bindings appear in the short help bar and, by subset, in full
	// help.
	Short
	// Full bindings appear in full help only.
	Full
)

// Binding wraps a live key.Binding with display-level (Vis) and
// displayed-key (Show) axes. It does not redeclare the binding's triggers,
// description, or enabled state — those are read live from the embedded
// binding via Keys(), Help().Desc, and Enabled().
type Binding struct {
	key.Binding

	// Show, when non-nil, is the subset of the embedded binding's Keys()
	// named in help. When nil it defaults to the full Keys(). Keys present
	// in Keys() but absent from Show (hidden vim aliases) route but never
	// appear in help output.
	Show []string

	// Vis selects which help bars display this binding. The zero value
	// RouteOnly hides it from both bars while still routing and claiming.
	Vis Vis
}

// displayKeys returns the binding's displayed-key set: Show when non-nil,
// otherwise the full Keys(). The returned slice is never mutated.
func (c Binding) displayKeys() []string {
	if c.Show != nil {
		return c.Show
	}
	return c.Keys()
}

// Group is both the unit of conflict resolution and one full-help display
// column. A Map may return several groups (e.g. navigation and actions as
// separate columns).
type Group = []Binding

// Map is the single contract every input layer implements. Keys()
// returns the layer's groups already aggregated, highest-priority first: a
// composite concatenates its focused child's Keys() (the child's full
// subtree) ahead of its own KeyMap's groups; a leaf returns its own
// groups. Priority is expressed solely by group order.
type Map interface {
	Keys() []Group
}
