package keymap

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Handles reports whether any enabled binding across all groups of
// child.Keys() (the child's aggregated subtree) matches msg, using each
// binding's complete Keys() regardless of Vis or Show. Because child.Keys()
// is aggregated, this delegates correctly at any nesting depth. Routing
// does not go through the help-resolution pipeline.
func Handles(child Map, msg tea.KeyPressMsg) bool {
	for _, g := range child.Keys() {
		for _, c := range g {
			if !c.Enabled() {
				continue
			}
			if key.Matches(msg, c.Binding) {
				return true
			}
		}
	}
	return false
}

// Resolve produces priority-merged, relabeled help from the flat
// priority-ordered group list. Processing groups left-to-right (highest
// priority first), it removes from each binding's displayed keys any key in
// the cumulative claim — the union of every earlier group's enabled
// bindings' complete Keys() (regardless of Vis; RouteOnly bindings still
// claim). A binding whose displayed keys become empty is dropped; survivors
// have their label rebuilt from the residual keys via render (description
// preserved) and retain their Vis. Empty groups are dropped and group
// order is preserved. Disabled bindings are skipped for both claiming and
// display. Resolve never mutates caller inputs.
func Resolve(render Render, groups ...Group) []Group {
	if render == nil {
		render = DefaultRender
	}
	claimed := map[string]struct{}{}
	out := make([]Group, 0, len(groups))

	for _, g := range groups {
		// Keys claimed by this group, applied to claimed only after the
		// whole group is processed so bindings within one group do not
		// shadow each other (a model's own groups are disjoint by
		// convention; cumulative-by-earlier-group is the spec rule).
		var groupClaim []string
		resolved := make(Group, 0, len(g))

		for _, c := range g {
			if !c.Enabled() {
				continue
			}
			groupClaim = append(groupClaim, c.Keys()...)

			residual := make([]string, 0, len(c.displayKeys()))
			for _, k := range c.displayKeys() {
				if _, taken := claimed[k]; !taken {
					residual = append(residual, k)
				}
			}
			if len(residual) == 0 {
				continue
			}
			resolved = append(resolved, Binding{
				Binding: key.NewBinding(
					key.WithKeys(c.Keys()...),
					key.WithHelp(render(residual), c.Help().Desc),
				),
				Show: residual,
				Vis:  c.Vis,
			})
		}

		for _, k := range groupClaim {
			claimed[k] = struct{}{}
		}
		if len(resolved) > 0 {
			out = append(out, resolved)
		}
	}
	return out
}

// ShortHelp projects a resolved set into the short help bar: the resolved
// groups flattened in priority order, keeping only Vis == Short bindings,
// emitted as relabeled key.Binding values. RouteOnly and dropped bindings
// are excluded.
func ShortHelp(groups []Group) []key.Binding {
	groups = Resolve(nil, groups...)
	var out []key.Binding
	for _, g := range groups {
		for _, c := range g {
			if c.Vis == Short {
				out = append(out, c.Binding)
			}
		}
	}
	return out
}

// FullHelp projects a resolved set into full help: one row per non-empty
// group (group boundaries preserved), keeping Vis ∈ {Short, Full} bindings,
// emitted as relabeled key.Binding values. RouteOnly and dropped bindings
// are excluded.
func FullHelp(groups []Group) [][]key.Binding {
	groups = Resolve(nil, groups...)
	var out [][]key.Binding
	for _, g := range groups {
		var row []key.Binding
		for _, c := range g {
			if c.Vis == Short || c.Vis == Full {
				row = append(row, c.Binding)
			}
		}
		if len(row) > 0 {
			out = append(out, row)
		}
	}
	return out
}
