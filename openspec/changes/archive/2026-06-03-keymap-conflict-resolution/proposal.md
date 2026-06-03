## Why

When two layers of the TUI bind the same key — a focused `selectfield` and its `form` both wanting `up`/`down`, or an overlay's `esc` colliding with an inner control's `esc` — there is no uniform rule for who wins, and no mechanism to keep help honest after a key is claimed. Routing is decided ad-hoc per layer (e.g. `form.go:121` tests a keypress against the focused field's `ShortHelp()`), which couples input routing to a presentation artifact, and help composition concatenates eagerly (`form.go:238`, `overlay.go:47`) so the same key appears under conflicting descriptions and bespoke dedup hacks (`overlay.go`'s `hasEsc`) accrete.

## What Changes

- Add an internal `keymap` package that is the single, stack-wide source of truth for keybinding ownership: one gesture routes to its highest-priority owner, and help is a priority-merged, relabeled projection.
- Introduce one uniform contract across **every** input layer (field, form, querybar, overlay, screen, app): `keymap.Map` with `Chords() []keymap.Group` (`Group = []Chord`). A `Chord` embeds the live `key.Binding` (so routing keys, `Enabled()`, and description ride along) and adds `Show` (displayed-key subset) and `Vis` (`RouteOnly` default / `Short` / `Full`). A composite produces its `Chords()` by concatenating its focused child's groups ahead of its own — the result is the active subtree flattened, highest-priority first — so no `Active`/`Stack`/`Collect` walker is needed.
- Routing uses the aggregated complete key set, eliminating the `ShortHelp`-as-routing-oracle coupling; a single `Handles(child, msg)` delegates correctly at any depth. Routing and help derive from the same `Chords()`, so they cannot diverge.
- Help resolves **once**: `Resolve(render, m.active.Chords()...)` subtracts from each group the cumulative union of all earlier (higher-priority) groups' complete keys, relabels survivors via an injectable glyph renderer, drops empties, and preserves order. Short and full bars are `Vis` projections of that single result.
- Make the help label a function of the residual key set via an injectable renderer (default glyph-join), because `key.Binding` stores a pre-rendered label that `SetKeys` does not update.
- **BREAKING (internal):** `screen.Screen` and `form.Field` adopt `keymap.Map`; eager help composition in `form` and `overlay` (including `overlay.go`'s `hasEsc` dedup) is replaced by `keymap.Resolve`. Free-text capture (`CapturingInput`) is unchanged and out of scope.

## Capabilities

### New Capabilities
- `keymap-resolution`: stack-wide keybinding ownership — a uniform `Map.Chords() []Group` contract aggregated by concatenation, a `Resolve` pass that produces priority-merged relabeled help in one go, a `Handles` predicate for incremental single-gesture routing at any depth, and a pluggable, non-mutating, enabled-aware label renderer.

### Modified Capabilities
- `form-field-toolkit`: fields declare a complete consumed keymap via `Bindings()`; the form routes to the focused field through it and contributes a group to resolution instead of concatenating help.
- `tui-view-stack`: the `Screen` contract adopts `keymap.Map`; overlay aggregates its inner's `Chords()` and routes/relabels through the resolver, retiring the `hasEsc` dedup.
- `tui-application`: the app footer resolves help via `Resolve(m.active.Chords()...)` and contributes its own global bindings as the lowest-priority groups.

## Impact

- New package: `tui/internal/keymap` (`Chord`, `Group`, `Vis`, `Map`, `Resolve`, `Handles`, default `Render` + glyph table, short/full projections).
- `tui/components/screen/screen.go`, `overlay.go`: `Screen` embeds `keymap.Map`; overlay aggregates inner `Chords()`; routing/help rewired; `hasEsc` removed.
- `tui/components/form/field.go`, `form.go`: `Field` embeds `keymap.Map`; form aggregates the focused field's `Chords()`; `Update` routing and help composition rewired.
- Concrete fields and leaf screens declare `Chords()` (notably `selectfield`/`radiofield` claiming `up`/`down`).
- `tui/app.go`: footer help via `Resolve`; app KeyMap becomes the lowest-priority groups.
- Unchanged: `CapturingInput`/free-text capture, `app.go` global-key *gating* semantics (still suppress global keys during free-text), third-party deps (`charm.land/bubbles/v2/key` only).
