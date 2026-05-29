## MODIFIED Requirements

### Requirement: Debounced live validation and apply
While focused, the query bar SHALL re-validate the current value after a configurable debounce interval following the last keystroke. The interval SHALL be provided at construction time. When the debounced parse succeeds, the query bar SHALL return an `ApplyMsg{Query: <trimmed value>}` via cmd so the parent reloads with the new filter; the query bar SHALL remain focused and SHALL NOT update its stored "applied query". When the debounced parse fails, the query bar SHALL store the `*ParseError` for inline highlighting and return a cmd yielding it as an `error` for the app error bar.

#### Scenario: Debounce fires after idle and the query is valid
- **WHEN** the user stops typing for the configured debounce interval and the current value parses cleanly
- **THEN** the query bar SHALL return an `ApplyMsg` carrying the trimmed current value
- **AND** the query bar SHALL remain focused
- **AND** the stored applied query SHALL NOT change

#### Scenario: Debounce fires after idle and the query is invalid
- **WHEN** the user stops typing for the configured debounce interval and the current value fails to parse
- **THEN** the parse error SHALL be stored for inline highlighting
- **AND** a cmd yielding the `*ParseError` as an `error` SHALL be returned
- **AND** no `ApplyMsg` SHALL be emitted

#### Scenario: Debounce resets on keystroke
- **WHEN** the user types another character before the debounce fires
- **THEN** the previous debounce SHALL be canceled and a new one started

### Requirement: Cancel on esc
Pressing esc while focused SHALL revert the query bar's text value to the last applied query, clear any parse error, blur, and return an `ApplyMsg{Query: <applied query>}` via cmd. Emitting `ApplyMsg` here (rather than a distinct cancel message) causes the parent to reload using the previously-applied filter, visibly undoing any live-previewed state.

#### Scenario: Cancel reverts and blurs
- **WHEN** the user presses esc while focused
- **THEN** the input value SHALL revert to the last applied query
- **AND** any parse error SHALL be cleared
- **AND** the query bar SHALL blur
- **AND** an `ApplyMsg` carrying the previously-applied query SHALL be returned

## REMOVED Requirements

### Requirement: CancelMsg type
**Reason**: Esc now emits `ApplyMsg` with the previously-applied query so the parent reloads back to its prior state in a single message. A distinct cancel signal is no longer needed; the `CancelMsg` type is removed from the package.

**Migration**: Parent screens that previously handled `case querybar.CancelMsg:` SHALL remove that case. The existing `case querybar.ApplyMsg:` handler covers both commit and revert, since esc now delivers an `ApplyMsg` carrying the prior query.
