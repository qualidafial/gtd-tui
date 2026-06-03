## 1. keymap package core

- [x] 1.1 Create `tui/internal/keymap` package with `Chord` (embeds `key.Binding`; adds `Show []string`, `Vis`), `Group = []Chord`, and `Vis` enum (`RouteOnly` zero value, then `Short`, `Full`)
- [x] 1.2 Implement `Render` type, default glyph table (`down→↓`, `up→↑`, `shift+tab→⇧tab`, …), and default renderer with raw-string fallback for unmapped keys
- [x] 1.3 Implement `Handles(child Map, msg tea.KeyPressMsg) bool` over the child's flattened groups using complete `Keys()`, skipping disabled chords
- [x] 1.4 Implement `Resolve(render Render, groups ...Group) []Group`: cumulative left-to-right claim from each chord's complete `Keys()` (incl. `RouteOnly`, skipping disabled), subtract from later chords' displayed `Show` keys, rebuild labels via `render`, drop empty chords and empty groups, preserve order, never mutate inputs
- [x] 1.5 Implement short-help and full-help projections from a single resolved set (`Vis==Short` flattened; `Vis∈{Short,Full}` as group rows), emitting `key.Binding`s with relabeled help
- [x] 1.6 Define the `Map` interface (`Chords() []Group`)

## 2. keymap package tests

- [x] 2.1 Table tests for `Resolve`: partial shadow relabels (`{tab,down}`→`tab`), full shadow drops chord, hidden alias never surfaced (`Keys{j,down}`/`Show{down}`), claim accumulates across multiple groups, disabled chords claim/show nothing, `RouteOnly` claims but never displays
- [x] 2.2 Tests for `Resolve` non-mutation (inputs unchanged across two calls) and group/order preservation
- [x] 2.3 Tests for short/full projections from one resolved set (Vis filtering, group boundaries, RouteOnly excluded)
- [x] 2.4 Tests for `Handles`: match by complete keys incl. hidden alias, disabled returns false, depth via aggregated child `Chords()`
- [x] 2.5 Tests for default renderer glyphs and fallback

## 3. Field contract

- [x] 3.1 Change `form.Field` interface: replace `ShortHelp`/`FullHelp` with `Chords() []keymap.Group`
- [x] 3.2 Implement `Chords()` on each concrete field (`inputfield`, `textfield`, `selectfield`, `radiofield`, `datefield`, `savefield`), wrapping existing `key.Binding`s; set `Vis` explicitly and `Show` only where a binding has hidden aliases
- [x] 3.3 `selectfield` and `radiofield` claim `up`/`down` via their `Chords()` so those route to the field
- [x] 3.4 Update each field's `keymap.go` to expose `Chords()` on the KeyMap struct

## 4. Form composition and routing

- [x] 4.1 Add `Chords()` to `form.KeyMap` (Next/Prev/Save) with appropriate `Vis`; mark navigation keys' `enter` alias appropriately (e.g. `RouteOnly` or `Show` excluding it)
- [x] 4.2 `form.Model.Chords()` = `slices.Concat(Focused().Chords(), m.KeyMap.Chords())`
- [x] 4.3 Rewire `form.Model.Update` routing: replace the `key.Matches(msg, field.ShortHelp()...)` check (`form.go:121`) with `keymap.Handles(focusedField, msg)` to protect field keys, matching own keys against `m.KeyMap.Chords()`
- [x] 4.4 Derive `form.Model.ShortHelp`/`FullHelp` from `keymap.Resolve(render, m.Chords()...)` projections; delete the eager `slices.Concat` composition (`form.go:240`, `form.go:249`)

## 5. Screen / overlay contract

- [x] 5.1 Change `screen.Screen` to embed `keymap.Map` instead of `help.KeyMap`
- [x] 5.2 Expose `Chords()` on `screen.KeyMap` (the overlay `Back`/esc binding)
- [x] 5.3 `overlay.Chords()` = `slices.Concat(inner.Chords(), o.KeyMap.Chords())`; remove `hasEsc`/`ShortHelp`/`FullHelp` dedup (`overlay.go:47-67`)
- [x] 5.4 Keep `overlay.Update` esc-to-dismiss gated by `CapturingInput`; verify esc subtraction now comes from `Resolve` (overlay esc dropped when inner claims esc)
- [x] 5.5 Update every concrete `Screen` (pages, status overlays, pickers, querybar host screens) to implement `Chords()` instead of `ShortHelp`/`FullHelp`

## 6. App footer

- [x] 6.1 Add `Chords()` to `tui.KeyMap` (Quit/Help)
- [x] 6.2 `app.Model.Chords()` = `slices.Concat(m.active.Chords(), m.KeyMap.Chords())`; render footer via `keymap.Resolve` projections fed to `help.Model`
- [x] 6.3 Keep `Help` binding disabled during `CapturingInput`; confirm `?` routing unchanged
- [x] 6.4 Update `app_test.go` `stubScreen` and any test screens to implement `Chords()`

## 7. Verification

- [x] 7.1 Run `go build ./...` and `go test ./...`; fix fallout from the interface change
- [x] 7.2 Manually verify in-app: focused `selectfield`/`radiofield` show `↑/↓` once (not duplicated under field-navigation), and `up`/`down` move the selection rather than fields
- [x] 7.3 Manually verify deep delegation: a key bound by both a field and the app routes to the field; esc on an overlay with a focused text field forwards (does not dismiss)
- [x] 7.4 Update `openspec/specs` impacted docs if any wording drift remains after implementation
