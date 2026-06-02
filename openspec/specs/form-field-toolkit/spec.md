# form-field-toolkit Specification

## Purpose

The form-field-toolkit capability provides an in-tree form framework for the gtd-tui TUI. It owns the `Form`/`Field` model, field navigation, validation, visibility, help integration, viewport-based scrolling, and a set of concrete field implementations (input, text, select, radio, date, save) that overlays compose. It replaces the prior dependency on `charm.land/huh`.

## Requirements

### Requirement: Form holds an ordered list of fields with a single focus
The toolkit SHALL expose a `Form` value that owns an ordered `[]Field` and tracks which field currently has focus. At any time exactly one field SHALL be focused (or none, after `Submit` succeeds and before reuse). The first field SHALL be focused after `Init`.

#### Scenario: Initial focus
- **WHEN** a `Form` is constructed with one or more fields and `Init` is called
- **THEN** the first field reports `Focused() == true` and every other field reports `Focused() == false`

#### Scenario: Tab advances focus
- **WHEN** the user presses the configured `Next` key (default `tab`) while focused on a non-last field
- **THEN** focus moves to the next field
- **AND** the previously focused field reports `Focused() == false`

#### Scenario: Shift+Tab retreats focus
- **WHEN** the user presses the configured `Prev` key (default `shift+tab`) while focused on a non-first field
- **THEN** focus moves to the previous field

### Requirement: Submit validates every field synchronously and reports first failure
`Form.Submit()` SHALL invoke `Validate()` on each field in declaration order. On the first validator that returns a non-nil error, `Submit` SHALL focus that field, cache the error on the field for `View` to render, and return `(ok=false, cmd=nil)` without invoking any side effects. If every validator passes, `Submit` SHALL return `(ok=true, cmd)` where `cmd` is any focus/blur cmd the form needs to apply. `Submit` MUST be synchronous: it MUST NOT dispatch messages through the bubbletea loop to drive validation.

#### Scenario: All fields valid
- **WHEN** every field's `Validate()` returns nil
- **THEN** `Submit` returns `ok = true`
- **AND** no field has a cached error

#### Scenario: First failure halts the walk
- **WHEN** a middle field's `Validate()` returns an error and a later field would also fail
- **THEN** `Submit` returns `ok = false`
- **AND** focus is on the first failing field
- **AND** the later field's validator was NOT called

#### Scenario: Submit from any focused field
- **WHEN** `Submit` is called while focus is on the last field
- **AND WHEN** `Submit` is called while focus is on the first field
- **THEN** the result is identical: every field is validated in declaration order

### Requirement: Ctrl+S triggers Submit from any field
The toolkit's default `KeyMap.Save` SHALL be `ctrl+s`. While a `Form` is focused on any field, pressing the configured `Save` key SHALL call `Submit` and surface the result to the caller via the returned cmd. The toolkit MUST NOT itself perform the overlay's save side-effect — it only validates and reports.

#### Scenario: Ctrl+S from a middle field with all valid
- **WHEN** the user presses `ctrl+s` while focused on a middle field and every field would validate
- **THEN** `Form.Update` returns a cmd that emits a `FormSubmittedMsg` (or equivalent signal) so the surrounding overlay can perform its save

#### Scenario: Ctrl+S with an invalid field
- **WHEN** the user presses `ctrl+s` and a field fails validation
- **THEN** focus jumps to the failing field
- **AND** no submitted-message is emitted

### Requirement: Hidden fields are excluded from navigation, rendering, validation, and submit
`form.Model` SHALL honor each field's `Visible(Values)` everywhere a field would otherwise participate:

- `View()` SHALL NOT render output for hidden fields.
- `Next()` and `Prev()` SHALL skip hidden fields when advancing focus.
- `Submit()` SHALL NOT call `Validate()` on hidden fields and SHALL NOT treat their cached errors as failures.
- `ShortHelp()` / `FullHelp()` SHALL NOT include hidden fields' bindings.

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
Both `Form` and `Field` SHALL implement `bubbles/v2/help.Model` (`ShortHelp() []key.Binding`, `FullHelp() [][]key.Binding`). `Form.ShortHelp()` SHALL return the form-level navigation/submit/cancel bindings composed with the currently focused field's bindings, so that as focus changes the help output updates without the surrounding screen having to know which field is focused. A screen that hosts a form SHOULD be able to delegate help with `return m.form.ShortHelp()`.

#### Scenario: Form help includes focused field bindings
- **WHEN** the focused field exposes a non-empty `ShortHelp()` (e.g. a `Text` field contributing `alt+enter`)
- **THEN** `Form.ShortHelp()` includes those bindings alongside the form's own (tab, shift+tab, ctrl+s)

#### Scenario: Form help updates when focus moves
- **WHEN** focus moves from a `Text` field to an `Input` field
- **THEN** `Form.ShortHelp()` no longer includes the `Text`-specific bindings
- **AND** does include any `Input`-specific bindings

### Requirement: Field interface contract
The toolkit SHALL expose a `Field` interface with at minimum `Key() string`, `Focus() (Field, tea.Cmd)`, `Blur() Field`, `Focused() bool`, `Visible(Values) bool`, `Update(tea.Msg) (Field, tea.Cmd)`, `View() string`, `Value() any`, `Validate() (Field, error)`, `SetWidth(int) Field`, `ShortHelp() []key.Binding`, and `FullHelp() [][]key.Binding`. Field interface methods that produce a new field state (`Focus`, `Blur`, `Update`, `Validate`, `SetWidth`) SHALL return a `Field` value rather than mutating in place. The default `Visible(Values)` for concrete field types SHALL return `true`; an `Option` (`WithVisible(func(form.Values) bool)`) lets callers wire conditional visibility.

The caller-supplied validator function passed via `WithValidator` MUST be a pure function of the field's value. `Field.Validate` invokes that validator and returns a new `Field` whose subsequent `View()` reflects the result (e.g. an inline error). The cached error in the returned field SHALL persist until the field's value changes via `Update`, at which point the cache SHALL be cleared so a stale error is not shown after the user has begun fixing the input.

`Field.Validate` is the only path through which an error becomes visible. Fields MUST NOT show validation errors purely as a function of their current value — errors appear only after an explicit `Validate` call (driven by `Submit` or by `Next` gating tab navigation).

#### Scenario: Mutating methods return new field values
- **WHEN** `field.Focus()` or `field.Blur()` or `field.Update(msg)` or `field.Validate()` is called
- **THEN** the returned `Field` is the new state; the receiver is not mutated in place

#### Scenario: Validate caches the result for View
- **WHEN** `Validate()` returns an error
- **THEN** the returned `Field`'s `View()` renders that error
- **AND** subsequent calls to `Validate()` on that field — without modifying the value — yield the same error

#### Scenario: Cached error clears when value changes
- **WHEN** a field has a cached error from a previous failing `Validate()`
- **AND WHEN** `Update` produces a value different from the value at the time `Validate` ran
- **THEN** the field's `View()` no longer renders the stale error

### Requirement: Tab gates on the current field's validator
`form.Model`'s `Next` action SHALL invoke the currently focused field's `Validate` before advancing focus. If the validator returns a non-nil error, focus SHALL stay on the current field and the field's `View()` SHALL display the error. `Prev` SHALL NOT gate — going backward must work even when the current field is invalid, so the user can revisit and fix earlier fields without being trapped.

#### Scenario: Tab refuses to leave an invalid field
- **WHEN** the focused field's validator would return a non-nil error
- **AND WHEN** the user presses `tab`
- **THEN** focus stays on that field
- **AND** the field's `View()` shows the validator's error

#### Scenario: Shift+Tab leaves an invalid field freely
- **WHEN** the focused field's validator would return a non-nil error
- **AND WHEN** the user presses `shift+tab`
- **THEN** focus moves to the previous visible field regardless

### Requirement: Field keys are unique and supplied at construction
Every concrete field's `New(...)` SHALL take the field's key as its first parameter (a non-empty string). `form.New(fields ...Field)` SHALL panic if two or more fields report the same `Key()` — uniqueness is enforced at construction time, not at runtime.

#### Scenario: Duplicate keys panic
- **WHEN** `form.New` is called with two fields whose `Key()` returns the same value
- **THEN** the call panics with a message naming the duplicate key

#### Scenario: Keys are how predicates name dependencies
- **WHEN** a `WithVisible(func(v form.Values) bool { return v.Get("kind") == "task" })` predicate runs on the form
- **THEN** `v.Get("kind")` returns the `Value()` of the field whose `Key()` is `"kind"`

### Requirement: Values is a progressive per-call snapshot
The `form.Values` interface SHALL expose at least `Get(key string) any`. `form.Model` SHALL produce `Values` snapshots progressively in declaration order: the snapshot supplied to field `i`'s `Visible` predicate SHALL contain only the `Value()` of visible fields with index `j < i`. A field MAY NOT read its own value through `Values`, and a hidden field's `Value()` MUST NOT appear in any snapshot supplied to a later field.

This rule keeps visibility a forward-only function of prior decisions, makes circular dependencies impossible by construction, and gives hidden fields no influence on later predicates.

`Values` MUST NOT carry any reference to specific `Field` instances — it is read-only, by-value data. Predicates MUST NOT depend on anything other than the `Values` they receive.

#### Scenario: Values reflect current field values
- **WHEN** a field with `Key() == "kind"` has `Value() == "task"`, and the form evaluates a later field's predicate
- **THEN** that predicate's `Values.Get("kind")` returns `"task"`

#### Scenario: Values updates as field values change
- **WHEN** the user changes a field's value via `Update`, and the form re-renders
- **THEN** the next predicate call sees a `Values` reflecting the new field value

#### Scenario: A field cannot read its own value
- **WHEN** field `c`'s `Visible` predicate calls `Values.Get("c")`
- **THEN** the result is `nil`

#### Scenario: Hidden field's value is excluded from later snapshots
- **WHEN** field `a` is hidden (its `Visible` returned false) and field `b` follows
- **AND WHEN** `b`'s predicate calls `Values.Get("a")`
- **THEN** the result is `nil` — hidden fields contribute nothing to subsequent snapshots

#### Scenario: Get on an unknown key returns nil
- **WHEN** a predicate calls `Values.Get` with a key not present in the form
- **THEN** the result is `nil`

### Requirement: Form scrolls tall content via a viewport
`form.Model` SHALL render through a `bubbles/v2/viewport.Model`. When the joined view of visible fields exceeds the form's available height, the form SHALL scroll automatically so that the currently focused field remains in view. `form.Model.Init` SHALL include `tea.RequestWindowSize` in its returned command so that the form is sized on its first render.

#### Scenario: Init requests a window size
- **WHEN** `Init` is called on a freshly-constructed form
- **THEN** the returned command, when run by the bubbletea runtime, requests a `tea.WindowSizeMsg`

#### Scenario: Focused field is brought into view
- **WHEN** the focused field's rendered range falls outside the current viewport offset
- **THEN** the form scrolls so the focused field's first line is visible

### Requirement: Window-size changes propagate to every field
On receipt of a `tea.WindowSizeMsg`, `form.Model.Update` SHALL invoke `SetWidth(width)` on every field (visible or not, so that previously-hidden fields are already sized when they become visible). The form SHALL also size its viewport from the same message.

#### Scenario: SetWidth fans out
- **WHEN** the form receives `tea.WindowSizeMsg{Width: 80, Height: 20}`
- **THEN** every field's `SetWidth(80)` is called exactly once

### Requirement: Concrete field subpackages
The toolkit SHALL provide concrete `Field` implementations as subpackages under `tui/components/form/`: `inputfield` (single-line text, wrapping `bubbles/v2/textinput`), `textfield` (multi-line text, wrapping `bubbles/v2/textarea`), `selectfield` (generic single-select backed by `bubbles/v2/list.Model`), `radiofield` (generic inline single-select for binary/small-N choices, equivalent to `huh.NewSelect().Inline(true)`), `datefield` (natural-time-parsed date/time, replacing `tui/components/date`), and `savefield` (a terminal `[ Save ]` button). Each subpackage exposes a `New(...)` constructor returning a `*Model` that implements `form.Field`. Each text-valued field SHALL accept an optional validator function via an `Option`.

#### Scenario: inputfield rejects empty when validator is supplied
- **WHEN** an `inputfield` is constructed with a validator that requires non-empty input
- **AND WHEN** the field's current value is the empty string
- **THEN** `Validate()` returns a non-nil error

#### Scenario: textfield preserves Alt+Enter newline binding
- **WHEN** the user presses `alt+enter` while focused on a `textfield`
- **THEN** a newline is inserted into the field's value

#### Scenario: textfield ignores plain Enter
- **WHEN** the user presses `enter` while focused on a `textfield`
- **THEN** the field's value SHALL NOT gain a newline
- **AND** the keypress propagates to the form (form-level navigation rules apply)

#### Scenario: radiofield navigates between options with left/right
- **WHEN** a `radiofield` with options `[A, B, C]` is focused on `A` and the user presses `right`
- **THEN** `B` becomes the selected value
- **AND WHEN** the user then presses `right` again
- **THEN** `C` becomes the selected value

#### Scenario: radiofield highlights the selected option when focused
- **WHEN** a `radiofield` is focused
- **THEN** its `View()` styles the currently selected option distinctly (accent color) from the unselected options

#### Scenario: selectfield cycles through options
- **WHEN** a `selectfield` with N options is focused and the user presses `down` (N-1) times
- **THEN** the last option is selected

#### Scenario: datefield snaps to absolute on blur
- **WHEN** a `datefield` contains a relative expression that parses to a non-nil time (e.g. "tomorrow at 3pm")
- **AND WHEN** the field loses focus
- **THEN** the displayed text is replaced by the absolute representation of the parsed time

#### Scenario: savefield always validates
- **WHEN** `Validate()` is called on a `savefield`
- **THEN** it returns nil

#### Scenario: Enter on savefield triggers form submit
- **WHEN** the user presses `enter` while focused on a `savefield` inside a form
- **THEN** the form runs the same Submit path as `ctrl+s`
- **AND** on success the form emits its submitted-message

### Requirement: Form rendering composes field views
`Form.View()` SHALL render each field's `View()` in order, vertically joined. A field's own `View()` SHALL include its label, its current value/edit affordance, and (when present) its cached error styled as an error message.

#### Scenario: Error is rendered beneath the failing field
- **WHEN** a field has a cached error after a failed `Validate()`
- **THEN** the field's `View()` includes the error text styled distinctly from the value

### Requirement: No huh dependency
The toolkit MUST NOT import any package from `charm.land/huh`. The `charm.land/huh/v2` module SHALL be removed from `go.mod` once the migration completes.

#### Scenario: huh is not in the module graph
- **WHEN** `go mod graph` is inspected after the migration completes
- **THEN** no `charm.land/huh` entries are present
