## Context

`form.Model.Update` decides whether a keypress belongs to the focused field or to form-level navigation by testing the keypress against the field's `ShortHelp()` (`form.go:121`). `form.Model.ShortHelp` then concatenates the field's bindings with the form's own (`form.go:238`). Two problems follow:

1. **Routing reads a presentation artifact.** `ShortHelp` is a curated, human-facing subset. Using it to route input means a key a field genuinely consumes but doesn't advertise is mis-routed, and a binding added for display reasons silently starts capturing input. The two concerns drift.
2. **Help double-lists conflicts.** When a `selectfield` claims `up`/`down` and the form's `Next = {tab, down, enter}` / `Prev = {shift+tab, up}` also list them, the help bar shows `↓` under both `select` and `next`. There is no mechanism to subtract the stolen key from the loser and relabel it.

Confirmed API constraint: `charm.land/bubbles/v2/key.Binding` couples a key set with a **pre-rendered** help label. `Help()` returns `{Key, Desc}` where `Key` (e.g. `"tab/↓"`) is the literal string passed to `WithHelp`; `SetKeys` does **not** update it. So removing a key from a binding cannot fix its label automatically — the label must be regenerated.

## Goals / Non-Goals

**Goals:**
- One internal package that is the single source of truth for "who owns this key" across priority-ordered layers.
- Both routing (one `tea.KeyPressMsg`) and help (relabeled `[]key.Binding`) derive from the same per-control binding data, so they cannot drift.
- Help labels become a function of the residual key set, so a stolen key disappears from the loser's label cleanly.
- Non-mutating: inputs (callers' stored `KeyMap` bindings) are never altered.

**Non-Goals:**
- Free-text capture. "Consume every printable rune" is not enumerable as bindings; it stays a boolean (`CapturingInput`) and is out of this package's scope. `app.go` global-key gating is unchanged.
- A `Handles(KeyPressMsg) bool` interface method on models. The enumerable consumed keymap is the contract; `Handles` is a derived free function over it (a predicate can be probed but not enumerated, so it cannot drive help subtraction).
- Replacing `ShortHelp`/`FullHelp` as the display interface. They remain; the form composes them through the resolver instead of by raw concatenation.

## Decisions

### 1. `Chord` *embeds* the live `key.Binding` — no parallel declarations
`key.Binding` couples all triggers into `Keys()` and a single opaque label string into `Help().Key`, with no key↔glyph mapping. That makes faithful partial relabeling impossible (you cannot reliably remove just `down` from `"tab/↓"`) and makes label-from-`Keys()` regeneration *wrong* — it would surface keys the author deliberately hid (vim alias `Keys {j, down}` shown as `↓` would leak `j`). The fix is **not** to replace `key.Binding` (that would force re-authoring every keymap and would snapshot away the live `Enabled()` bit) but to wrap it:

```go
type Vis uint8 // increasing visibility; zero value is RouteOnly
const (
    RouteOnly Vis = iota // routes & claims, never displayed
    Short                // short bar (and full help)
    Full                 // full help only
)
type Chord struct {
    key.Binding          // embedded: Keys(), Enabled(), Help().Desc ride along, live
    Show  []string       // axis B: displayed-key subset; nil ⇒ defaults to Keys()
    Vis   Vis            // axis A: RouteOnly (default) | Short | Full
}
```

Two orthogonal display axes that a curated `ShortHelp`/`FullHelp` pair conflates: **A** = which chords appear (`short ⊆ full ⊆ all`, via `Vis`); **B** = which keys of a shown chord are named (via `Show`). Routing and claiming use *all* enabled chords' complete `Keys()` regardless of `Vis`; short/full help are filtered projections of one `Resolve` pass (`Vis==Short` for the bar; `Vis∈{Short,Full}` per-group rows for full). `RouteOnly` is the **default**, so visibility is explicit: a plain `Chord{Binding: b}` routes and claims but is shown nowhere until annotated. This is strictly better than today, where hiding a key from the bar also stopped it routing.

Consequences:
- **No second source of truth.** Existing `KeyMap`s stay `key.Binding`. A layer's `Chords()` is a thin projection — `{Binding: m.KeyMap.Next}` — the same shape as the `ShortHelp()` it would write anyway. Bindings are not redeclared.
- **Bubbles-internal bindings are never wrapped.** Components (textinput, list, viewport) are leaf layers consumed via their own `ShortHelp`/`FullHelp`; the resolver only re-renders the layers actually placed in conflict (field/form/overlay/app), so it never reaches into third-party keymaps.
- **`Show` is opt-in.** Set it only on a binding that both has hidden aliases *and* participates in a conflict (the vim case). Otherwise `Show` is `nil`, defaults to `Keys()`, and relabeling only ever *removes* an already-shown key — never adds a hidden one.

### 2. Three projections, distinct roles — claim is always the full key set
- **Routing**: `Chord.Keys` (complete). Used by `Handles`. Never relabeled, never goes through `Resolve`.
- **Help (short / full)**: built from `Chord.Show`. This is what `Resolve` edits.
- **Claim** (what a higher layer subtracts from a lower layer's help): the higher layer's complete `Chord.Keys`, **not** its `Show`. Rationale: if a higher layer *routes* a key (even one it doesn't display), pressing that key does the higher-layer thing — so showing it on a lower layer is a lie and must be suppressed. `Resolve` therefore takes, per layer, the full keys (claim) and the displayed chords (surface); it is run once per help variant.

*Alternative considered — whole-binding suppression* (drop a lower chord only if **all** its keys are shadowed): rejected because the real case is *partial* overlap (`Next = {tab, down, enter}, Show {tab, down}` vs field `{down}`); whole-chord suppression would leave `down` double-listed. Partial subtraction over `Show` is required.

### 3. Label is rebuilt from residual `Show` via an injectable glyph renderer
After subtraction, a surviving chord's label is `render(remainingShow) + " " + Desc`, where `type Render func(keys []string) string` defaults to a glyph-join over a small table (`down→↓`, `up→↑`, `shift+tab→⇧tab`, …; unmapped keys fall back to the raw string). Because `render` only ever sees keys that were in `Show`, no hidden trigger can leak. A chord is dropped from help only when `Show` becomes empty (even if `Keys` still carries routing aliases). `Desc` is always preserved.

### 4. One contract — `Chords() []Group` — composed by concatenation, no walker
A *group* (`type Group = []Chord`) is both the unit of conflict resolution and one full-help column; a model may contribute several (today's `FullHelp() [][]key.Binding` already does). Priority is just **order**, so there is no separate layer-hierarchy type and no `Active()`/`Collect`/`Stack`. A composite aggregates by concatenating its focused child's groups (already a full subtree) ahead of its own:

```go
type Group = []Chord
type Map   interface { Chords() []Group }   // aggregated: child subtree first, then own
// composite: return slices.Concat(m.Focused().Chords(), m.KeyMap.Chords())
// leaf:      return m.KeyMap.Chords()
```

Because `Chords()` is aggregated, a single flat `Handles(child, msg)` correctly delegates at **any depth** — pressing `?` with `app › overlay › form › field` resolves because `overlay.Chords()` already carries the field's `?`. (Own-only `Chords()` + `Active()` would have needed a *recursive* `Handles` to achieve the same; concatenation gets it for free, which is the decisive reason to drop the walker.)

Routing is **incremental**: each composite's `Update` fires its own key only when its child subtree doesn't claim it — matching `m.KeyMap.Chords()` (own) for the action and `Handles(child, msg)` (aggregated subtree) to protect the child. Help is resolved **once**: `Resolve(m.active.Chords()...)` walks the flat group list left-to-right, subtracting from each group's displayed keys the **cumulative** union of all earlier groups' complete `Keys()` (regardless of `Vis`), relabels survivors, drops empties. Cumulative-by-order equals per-layer claiming because a model's own groups are disjoint (same-key-in-two-columns is an authoring error) — flattening even yields free intra-model dedup. Group order + `Vis==Short` gives the flat short bar; the surviving groups are the full-help columns.

The aggregated `Chords()` consumed by parents/`Resolve` and the own `m.KeyMap.Chords()` consumed by the model's own routing decision are two objects already in hand (the child model vs the KeyMap), so no second interface method is required.

```go
func Handles(child Map, msg tea.KeyPressMsg) bool   // key.Matches over the child's (aggregated) groups
func Resolve(render Render, groups ...Group) []Group
```

### 5. Never mutate inputs
`Chord` carries plain slices and `Resolve` builds new `Chord`/`key.Binding` values; it never calls `key.Binding`'s pointer-receiver mutators (`SetKeys`, `SetEnabled`, `Unbind`) on a caller's binding (a struct copy can still share the backing slice). This keeps a layer's stored `Chords()`/`KeyMap` intact across renders.

### 6. Enabled-awareness
A disabled chord claims nothing and displays nothing this frame. `Resolve` and `Handles` both skip disabled chords, matching `key.Matches`'s enabled filtering so help subtraction and routing agree.

## Risks / Trade-offs

- **Loss of hand-authored key glyphs on shareable bindings** → Mitigated by keeping `Desc` authored and centralizing glyphs in one table; only bindings that participate in conflicts lose bespoke key strings, and those are exactly the ones that must be regenerated to stay honest.
- **Glyph table drift / unknown keys** → `render` falls back to the raw key string for unmapped keys, so an unmapped key degrades to `"ctrl+x"` rather than breaking.
- **Per-render cost** → `Resolve` runs on each help render and routing runs per keypress; both are O(keys) over a handful of bindings, negligible. Routing must not allocate per keypress beyond what `key.Matches` already does.
- **Reintroducing the conflation if `ShortHelp` is reused as the consumed source** → Mitigated by making the consumed keymap a separate accessor; the spec forbids routing off `ShortHelp`.
- **Free-text fields** → Out of scope by design; their runes never enter a layer, and `CapturingInput` continues to gate global keys. A reviewer might expect this package to subsume `CapturingInput`; the non-goal is explicit.

## Migration Plan

1. Land `tui/internal/keymap` with `Resolve`, `Handles`, default `Render`, glyph table, and tests (pure package, no TUI deps beyond `key`).
2. Add the consumed-keymap accessor to `Field` and implement it on each concrete field.
3. Rewire `form.Model.Update` routing and `ShortHelp`/`FullHelp` through the package.
4. Verify `selectfield`/`radiofield` `up`/`down` no longer double-list and route to the field, not field navigation, while focused.

Rollback: the package is additive; reverting steps 3–2 restores the `ShortHelp`-based routing without touching the new package.

## Open Questions

- Layer model surface: variadic `[]key.Binding` layers vs a named `Layer` type. Leaning variadic for now; a named type can wrap it later if overlay/app layers join.
- Whether the overlay and app keymaps should immediately become layers in the same `Resolve` call, or whether this change stays scoped to form↔field and the broader stack is a follow-up. Current scope: form↔field only.
