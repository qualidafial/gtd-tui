## MODIFIED Requirements

### Requirement: Hidden fields are excluded from navigation, rendering, validation, and submit
`form.Model` SHALL honor each field's `Visible(Values)` everywhere a field would otherwise participate:

- `View()` SHALL NOT render output for hidden fields.
- `Next()` and `Prev()` SHALL skip hidden fields when advancing focus.
- `Submit()` SHALL NOT call `Validate()` on hidden fields and SHALL NOT treat their cached errors as failures.
- `Keys()` SHALL NOT include hidden fields' bindings. Because `Form.Keys()` aggregates only the focused field's bindings and a hidden field cannot hold focus, hidden fields contribute nothing to routing or help.

If focus is on a field whose `Visible()` flips to false between updates, `form.Model` SHALL move focus to the next visible field on the following `Update`. Visibility predicates are evaluated lazily (on each render and each navigation/submit call) — callers MUST NOT need to notify the form when an upstream field's value changes.

#### Scenario: Hidden field skipped during tab navigation
- **WHEN** a form has fields `[A, B, C]` and `B.Visible(values)` returns `false` for the current snapshot
- **AND WHEN** focus is on `A` and the user presses `tab`
- **THEN** focus moves to `C`, not `B`

#### Scenario: Hidden field excluded from Submit
- **WHEN** a form has fields `[A, B]`, `A.Validate()` returns nil, `B.Validate()` would return an error, but `B.Visible(values)` returns `false`
- **THEN** `Submit()` returns `ok = true`
- **AND** `B.Validate()` is NOT called

#### Scenario: Hidden field not rendered
- **WHEN** a field reports `Visible(values) == false` for the current snapshot
- **THEN** that field's `View()` output does not appear in the form's rendered output

#### Scenario: Focused field becomes hidden
- **WHEN** focus is on field `X` and an `Update` changes another field's value such that `X.Visible(values)` now returns `false`
- **THEN** on the same or following `Update`, focus moves to the next visible field

### Requirement: Form and Field expose help bindings reactively
Both `Form` and `Field` SHALL implement `keymap.Map` (`Keys() []keymap.Group`). A `Field` SHALL return its own groups. `Form.Keys()` SHALL aggregate by concatenating the focused field's `Keys()` (the field's full subtree) ahead of its own KeyMap groups, so the result is the active form subtree flattened, highest-priority first. Cross-layer help composition — resolving key conflicts across the aggregated groups and relabeling — SHALL be performed by the `keymap` package's `Resolve`, not by the form. A field that consumes a key the form also binds (e.g. `up`/`down`) SHALL claim it so the form-level navigation binding is subtracted from help while that field is focused. `Form` SHALL still satisfy `bubbles/v2/help.Model` for rendering, with its `ShortHelp`/`FullHelp` derived from the resolved bindings (short bar = `Vis == Short`; full = `Vis ∈ {Short, Full}`).

#### Scenario: Form bindings aggregate child then own
- **WHEN** `Form.Keys()` is read while a field is focused
- **THEN** the focused field's groups appear first
- **AND** the form's own navigation/submit groups follow

#### Scenario: Focused field claim removes the duplicate from form help
- **WHEN** the focused field claims `down` and the form binds `down` as part of next-field
- **THEN** the resolved help shows the field's `down` action
- **AND** the form's next-field binding is relabeled without `down` (e.g. `tab` remains)

#### Scenario: Help updates when focus moves
- **WHEN** focus moves from a field that claims `up`/`down` to one that does not
- **THEN** the resolved help no longer subtracts `up`/`down` from the form-level bindings

### Requirement: Field interface contract
The toolkit SHALL expose a `Field` interface with at minimum `Key() string`, `Focus() (Field, tea.Cmd)`, `Blur() Field`, `Focused() bool`, `Visible(Values) bool`, `Update(tea.Msg) (Field, tea.Cmd)`, `View() string`, `Value() any`, `Validate() (Field, error)`, `SetWidth(int) Field`, and `Keys() []keymap.Group`. The `Keys()` method SHALL return the field's complete consumed bindings as the single source for routing and help; a field SHALL NOT expose a separate curated routing list. Field interface methods that produce a new field state (`Focus`, `Blur`, `Update`, `Validate`, `SetWidth`) SHALL return a `Field` value rather than mutating in place. The default `Visible(Values)` for concrete field types SHALL return `true`; an `Option` (`WithVisible(func(form.Values) bool)`) lets callers wire conditional visibility.

The caller-supplied validator function passed via `WithValidator` MUST be a pure function of the field's value. `Field.Validate` invokes that validator and returns a new `Field` whose subsequent `View()` reflects the result (e.g. an inline error). The cached error in the returned field SHALL persist until the field's value changes via `Update`, at which point the cache SHALL be cleared so a stale error is not shown after the user has begun fixing the input.

`Field.Validate` is the only path through which an error becomes visible. Fields MUST NOT show validation errors purely as a function of their current value — errors appear only after an explicit `Validate` call (driven by `Submit` or by `Next` gating tab navigation).

#### Scenario: Field declares consumed bindings via Keys
- **WHEN** a `selectfield`/`radiofield` is focused
- **THEN** its `Keys()` includes `up`/`down` so they route to the field, not field navigation

#### Scenario: Mutating methods return new field values
- **WHEN** `field.Focus()` or `field.Blur()` or `field.Update(msg)` or `field.Validate()` is called
- **THEN** the returned `Field` is the new state; the receiver is not mutated in place

#### Scenario: Validate caches the result for View
- **WHEN** `Validate()` returns an error
- **THEN** the returned `Field`'s `View()` renders that error
- **AND** subsequent calls to `Validate()` on that field — without modifying the value — yield the same error
