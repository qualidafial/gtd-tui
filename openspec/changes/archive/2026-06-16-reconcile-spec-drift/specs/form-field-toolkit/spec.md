## ADDED Requirements

### Requirement: Select options populated after construction
A `selectfield` SHALL support populating its options after construction so a form can be built up front and its selects filled once their data loads. The toolkit SHALL provide:

- `selectfield.Model.SetOptions([]Option[T]) Model[T]` — replaces the field's options, auto-sizing the list unless a height was pinned via `WithHeight`.
- A `WithHideWhenEmpty[T]()` field option — while the field has no real (non-`WithNone`) options its `Visible` SHALL return `false`; it composes with `WithVisible` (the field shows only when it has options AND the predicate passes).
- `form.Model.UpdateField(key string, fn func(Field) Field) Model` — applies an in-place edit to the field with the given key without rebuilding the form, leaving other fields untouched.

#### Scenario: SetOptions replaces a select's options
- **WHEN** `SetOptions` is called on a `selectfield` with a new set of options
- **THEN** the field's selectable options become exactly that set

#### Scenario: WithHideWhenEmpty hides the field until options arrive
- **WHEN** a `selectfield` built with `WithHideWhenEmpty` has no real options
- **THEN** its `Visible` returns `false`
- **AND WHEN** options are later supplied via `SetOptions`
- **THEN** its `Visible` returns `true` (subject to any `WithVisible` predicate)

#### Scenario: UpdateField edits one field in place by key
- **WHEN** `form.UpdateField` is called with an existing field's key and a mutator
- **THEN** that field is replaced by the mutator's result and the other fields are unchanged
