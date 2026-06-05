## MODIFIED Requirements

### Requirement: Binding wraps a live key.Binding with display-level and displayed-key axes
The `keymap` package SHALL define a `Binding` type that embeds `charm.land/bubbles/v2/key.Binding` and adds two orthogonal display controls: `Show []string` (which of the binding's keys are named in help) and `Vis` (in which help bars the binding appears). `Binding` SHALL NOT redeclare the binding's triggers, description, or enabled state — those SHALL be read live from the embedded `key.Binding` (`Keys()`, `Help().Desc`, `Enabled()`).

`Show`, when non-nil, SHALL be a subset of the embedded binding's `Keys()`; when nil it SHALL default to the full `Keys()`. Keys present in `Keys()` but absent from `Show` (hidden vim aliases) SHALL route but SHALL NOT appear in any help output.

`Vis` SHALL be an enum declared in increasing order of visibility with `RouteOnly` as the zero value: `RouteOnly` (routes and claims, never displayed), `Short` (appears in the short bar and, by subset, in full help), `Full` (appears in full help only). `Vis` SHALL affect display only: it SHALL NOT affect routing or claiming. Because `RouteOnly` is the default, a plain `Binding{Binding: b}` SHALL route and claim but SHALL NOT appear in any help bar until its `Vis` is set explicitly.

#### Scenario: Show defaults to all keys
- **WHEN** a `Binding` is constructed with `Show` nil over a binding with `Keys() == {"tab","down"}`
- **THEN** its displayed keys are `{"tab","down"}`

#### Scenario: Hidden alias routes but is never displayed
- **WHEN** a `Binding` wraps a binding with `Keys() == {"j","down"}` and `Show == {"down"}`
- **THEN** the binding's routing keys include `"j"` and `"down"`
- **AND** no help output produced from the binding ever contains `"j"`

#### Scenario: RouteOnly binding routes and claims but is hidden from both bars
- **WHEN** a `Binding` has `Vis == RouteOnly` and `Keys() == {"enter"}`
- **THEN** it matches in `Handles` and suppresses a lower layer's `enter` in help
- **AND** it never appears in short or full help output

#### Scenario: Enabled state is read live from the embedded binding
- **WHEN** the embedded `key.Binding` is later disabled via `SetEnabled(false)` on the owning KeyMap
- **THEN** a `Binding` wrapping that binding reports the disabled state without the wrapper having to be rebuilt

### Requirement: Group is the unit; composites aggregate by concatenation
The package SHALL define `type Group = []Binding` as both the unit of conflict resolution and one full-help display column. It SHALL define a single `Map` interface with `Keys() []Group`. A `Map` MAY return multiple groups (e.g. navigation and actions as separate columns). There SHALL be no separate layer-hierarchy type and no `Active`/`Collect`/`Stack`: a composite SHALL produce its `Keys()` by concatenating its focused child's `Keys()` (already the child's full subtree) ahead of its own KeyMap's groups, so the returned slice is the entire active subtree flattened, highest-priority first. Leaf layers SHALL return their own KeyMap's groups. Priority SHALL be expressed solely by group order (earlier = higher priority).

Because `Keys()` is aggregated, the result consumed by a parent or by `Resolve` carries the whole subtree's bindings. A model's own KeyMap groups (`m.KeyMap.Keys()`) remain separately available for the model's own routing decision; no second interface method is required.

#### Scenario: Composite aggregates child subtree then own
- **WHEN** a form with a focused field returns `Keys()`
- **THEN** the focused field's groups appear first (higher priority)
- **AND** the form's own groups follow

#### Scenario: Aggregation spans depth
- **WHEN** the stack is overlay → form → focused field and `overlay.Keys()` is read
- **THEN** the result contains the field's, the form's, and the overlay's groups, in that order

#### Scenario: One model may contribute several groups
- **WHEN** a model's KeyMap returns separate groups (e.g. navigation and actions)
- **THEN** each appears as its own full-help column at that model's priority position

### Requirement: Single-gesture routing is incremental and reads the complete key set
The package SHALL provide `Handles(child Map, msg tea.KeyPressMsg) bool` reporting whether any enabled binding across all groups of `child.Keys()` (the child's aggregated subtree) matches `msg`, using the bindings' complete `Keys()` (not `Show`, regardless of `Vis`). Routing SHALL be performed incrementally by each composite: before firing its own binding, a composite SHALL defer to its focused child when `Handles(child, msg)` is true. Because `child.Keys()` is aggregated, this delegates correctly at any nesting depth. Routing SHALL NOT go through the help-resolution pipeline.

#### Scenario: Composite defers a claimed key to its child
- **WHEN** a focused field claims `down` and the parent form also binds `down` to next-field
- **AND** the user presses `down`
- **THEN** the form forwards the keypress to the field
- **AND** does not advance field focus

#### Scenario: Deep delegation through intermediate layers
- **WHEN** the stack is app → overlay → form → field, the field claims `?`, and the overlay does not
- **AND** the user presses `?`
- **THEN** the app forwards rather than toggling help, because the overlay's aggregated `Keys()` carries the field's `?`

#### Scenario: Composite handles an unclaimed key locally
- **WHEN** the focused field does not claim `tab` and the form binds `tab` to next-field
- **AND** the user presses `tab`
- **THEN** the form advances field focus

#### Scenario: Disabled child binding is not handled
- **WHEN** a child binding matching the keypress is disabled
- **THEN** `Handles` returns false and the parent may handle the key

### Requirement: Resolve subtracts cumulative higher-priority claims and relabels survivors
The package SHALL provide `Resolve(render Render, groups ...Group) []Group` for help display, operating once over the flat priority-ordered group list. Processing groups left-to-right (highest priority first), `Resolve` SHALL remove from each binding's displayed keys any key in the **cumulative claim** — the union of every earlier group's enabled bindings' complete `Keys()` (not `Show`, and regardless of `Vis`; `RouteOnly` bindings still claim). A binding whose displayed keys become empty SHALL be dropped. Surviving bindings SHALL have their label rebuilt from the residual displayed keys via `render`, with the description preserved from the embedded binding, and SHALL retain their `Vis` for later projection. A group left with no surviving bindings SHALL be dropped. Group order SHALL be preserved. `Resolve` SHALL skip disabled bindings for both claiming and display.

#### Scenario: Partially shadowed binding is relabeled, not dropped
- **WHEN** an earlier group claims `down` and a later group has a binding with displayed keys `{tab, down}` described `next`
- **THEN** the resolved binding displays `{tab}` described `next`
- **AND** is not dropped

#### Scenario: Fully shadowed binding is dropped from help
- **WHEN** an earlier group claims `esc` and a later group's only displayed key for a binding is `esc`
- **THEN** that binding is absent from the resolved output

#### Scenario: Claim uses complete keys even when the higher layer hides the key
- **WHEN** an earlier group routes `down` but does not display it (`down` in `Keys()` but not `Show`)
- **AND** a later group displays `down`
- **THEN** the later group's `down` is removed from help

#### Scenario: Claim accumulates across all earlier groups
- **WHEN** the field contributes `down` and the form (later) and overlay (later still) each display `down`
- **THEN** both the form's and the overlay's `down` are removed from help

#### Scenario: Group order preserved
- **WHEN** `Resolve` is given groups in highest-first order
- **THEN** the returned groups are in the same order

### Requirement: Label rendering is injectable and cannot surface hidden keys
The package SHALL define `type Render func(keys []string) string` and a default renderer that joins per-key glyphs from a central table (e.g. `down→↓`, `up→↑`, `shift+tab→⇧tab`), falling back to the raw key string for unmapped keys. `render` SHALL only ever receive keys that were in a binding's displayed set, so a hidden routing alias can never appear in a rendered label. Hosts MAY supply an alternative `Render`.

#### Scenario: Default renderer maps known glyphs
- **WHEN** the default renderer is given `{"shift+tab","up"}`
- **THEN** it produces a label containing `⇧tab` and `↑`

#### Scenario: Unmapped key falls back to raw string
- **WHEN** the default renderer is given a key with no glyph entry (e.g. `"ctrl+x"`)
- **THEN** the label contains `ctrl+x`

### Requirement: Resolution never mutates caller inputs
`Resolve` SHALL NOT call any pointer-receiver mutator (`SetKeys`, `SetEnabled`, `Unbind`) on a caller's `key.Binding`, and SHALL NOT alter any caller-owned slice. Relabeled results SHALL be produced as new `Binding`/`key.Binding` values so a layer's stored KeyMap is unchanged across repeated calls.

#### Scenario: Repeated resolution is stable
- **WHEN** `Resolve` is invoked twice on the same layer stack
- **THEN** both calls return equivalent output
- **AND** the layers' stored bindings are unchanged after each call

### Requirement: Short and full help are filtered projections of one resolved set
The package SHALL produce both help bars from a single `Resolve` result, without resolving twice. It SHALL provide a short-help projection that flattens the resolved groups in priority order and keeps only bindings with `Vis == Short`, and a full-help projection that returns the resolved groups as rows (preserving group boundaries) keeping bindings with `Vis ∈ {Short, Full}`. Both projections SHALL emit `key.Binding` values carrying the relabeled help, and SHALL exclude bindings dropped during resolution and bindings marked `RouteOnly`.

#### Scenario: Short help is a flat priority-ordered subset
- **WHEN** the resolved set contains bindings with mixed `Vis`
- **THEN** the short-help projection contains only `Vis == Short` bindings, flattened highest-priority first

#### Scenario: Full help preserves group boundaries
- **WHEN** the resolved set spans field, form, and overlay groups
- **THEN** the full-help projection returns one row per non-empty group, keeping `Short` and `Full` bindings

#### Scenario: One resolution drives both bars
- **WHEN** both the short and full projections are requested for the same stack
- **THEN** they derive from a single `Resolve` pass over the aggregated group list
- **AND** a key claimed by a higher-priority group is absent from both bars
