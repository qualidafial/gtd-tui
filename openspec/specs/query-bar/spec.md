# query-bar Specification

## Purpose
Defines the reusable `querybar` component: a single-line text input with focus/blur lifecycle, debounced validation, parse-error highlighting, and apply/cancel messaging used by list screens that filter via compact query strings.

## Requirements

### Requirement: Reusable query bar component
The `querybar` package SHALL provide a `Model` that wraps a `textinput.Model` with focus/blur management, debounced validation, parse-error state, and single-line rendering. The component SHALL accept a `ValidateFunc func(string) *ParseError` at construction time for query validation.

#### Scenario: Construction with validate function
- **WHEN** a query bar is created with a validate function, prompt, and placeholder
- **THEN** the model SHALL store the validate function and configure the underlying textinput

### Requirement: Focus and blur
The query bar SHALL support `Focus()` and `Blur()` methods. `CapturingInput()` SHALL return true when focused. On focus, if the current value is non-empty, a trailing space SHALL be appended so that new typing becomes a separate term. On blur, leading and trailing whitespace SHALL be trimmed from the value.

#### Scenario: Focus activates text input
- **WHEN** `Focus()` is called
- **THEN** the underlying textinput SHALL be focused
- **AND** `CapturingInput()` SHALL return true

#### Scenario: Focus appends trailing space
- **WHEN** `Focus()` is called and the current value is `status:open`
- **THEN** the textinput value SHALL become `status:open ` (with trailing space)
- **AND** the cursor SHALL be at the end

#### Scenario: Focus on empty value does not add space
- **WHEN** `Focus()` is called and the current value is empty
- **THEN** the textinput value SHALL remain empty

#### Scenario: Blur deactivates text input
- **WHEN** `Blur()` is called
- **THEN** the underlying textinput SHALL be blurred
- **AND** `CapturingInput()` SHALL return false

### Requirement: Apply on enter
Pressing enter while focused SHALL trim leading/trailing whitespace from the current value, then validate it. On success, the query bar SHALL blur, store the trimmed value as the applied query, and return a `querybar.ApplyMsg{Query string}` via cmd. The parent handles this message to parse the query into its typed filter and reload. On validation failure, the query bar SHALL store the `*ParseError` for inline highlighting and return a cmd yielding it as an `error` for the app error bar.

#### Scenario: Successful apply
- **WHEN** the user presses enter and validation succeeds
- **THEN** the query bar SHALL blur
- **AND** return an `ApplyMsg` with the current query string
- **AND** store the value as the applied query

#### Scenario: Failed apply
- **WHEN** the user presses enter and validation returns a non-nil `*ParseError`
- **THEN** the query bar SHALL store the parse error for inline highlighting
- **AND** return a cmd yielding an `error` for the app error bar
- **AND** the query bar SHALL remain focused

### Requirement: Cancel on esc
Pressing esc while focused SHALL revert the query bar's text value to the last applied query, clear any parse error, blur, and return an `ApplyMsg{Query: <applied query>}` via cmd. Emitting `ApplyMsg` here (rather than a distinct cancel message) causes the parent to reload using the previously-applied filter, visibly undoing any live-previewed state.

#### Scenario: Cancel reverts and blurs
- **WHEN** the user presses esc while focused
- **THEN** the input value SHALL revert to the last applied query
- **AND** any parse error SHALL be cleared
- **AND** the query bar SHALL blur
- **AND** an `ApplyMsg` carrying the previously-applied query SHALL be returned

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

### Requirement: Single-line rendering
The query bar SHALL always render as a single line. When no parse error is present, it SHALL delegate to `textinput.View()`. When a parse error is present, it SHALL post-process the `textinput.View()` output using `ansi.Cut` to slice the rendered text around the error's `[Start, End)` rune range (offset by prompt width), apply red foreground and underline styling to the offending segment, and concatenate the segments.

#### Scenario: Normal rendering
- **WHEN** no parse error is present
- **THEN** the view SHALL be the textinput's single-line output

#### Scenario: Error highlighting
- **WHEN** a parse error marks runes 5-12 as offending
- **THEN** the view SHALL show the textinput output with runes 5-12 (relative to the text value, offset by prompt width) styled with red foreground and underline

### Requirement: ParseError type
The `querybar` package SHALL define a `ParseError` struct with `Message string`, `Start int`, and `End int` fields. `Start` and `End` are rune offsets into the query string marking the `[Start, End)` range of the offending token. `ParseError` SHALL implement the `error` interface.

#### Scenario: ParseError is an error
- **WHEN** a `*ParseError` is returned from a validate function
- **THEN** it SHALL satisfy the `error` interface with `Message` as the error string
