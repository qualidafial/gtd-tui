## ADDED Requirements

### Requirement: Enter on the last visible field submits
When the focused field is the last visible field and does not itself claim the Enter key, pressing Enter SHALL run the form's Submit path (identical to the `Save` key), and on success the form SHALL emit its submitted-message. When the focused field claims Enter (its `Keys()` include Enter), the keypress SHALL route to that field instead, unchanged. When a later visible field exists, Enter SHALL continue to advance focus (gated by the current field's validator) rather than submit. This makes any non-Enter-claiming terminal field — for example an inline `radiofield` placed last — act as the form's submit affordance without bespoke wiring.

#### Scenario: Enter on a non-claiming last field submits
- **WHEN** focus is on the last visible field, that field does not claim Enter, and the user presses Enter
- **THEN** the form runs the same Submit path as the `Save` key
- **AND** on success the form emits its submitted-message

#### Scenario: Enter still advances when a later field exists
- **WHEN** focus is on a non-last visible field and the user presses Enter
- **THEN** focus advances to the next visible field (subject to the current field's validator) and the form does not submit

#### Scenario: A claiming last field keeps Enter
- **WHEN** focus is on the last visible field and that field claims Enter (e.g. a multi-line field with an Enter binding)
- **THEN** the keypress routes to the field and the form does not submit on its behalf

## MODIFIED Requirements

### Requirement: Concrete field subpackages
The toolkit SHALL provide concrete `Field` implementations as subpackages under `tui/components/form/`: `inputfield` (single-line text, wrapping `bubbles/v2/textinput`), `textfield` (multi-line text, wrapping `bubbles/v2/textarea`), `selectfield` (generic single-select backed by `bubbles/v2/list.Model`), `radiofield` (generic inline single-select for binary/small-N choices, equivalent to `huh.NewSelect().Inline(true)`), `datefield` (natural-time-parsed date/time, replacing `tui/components/date`), and `savefield` (a terminal `[ Save ]` button). Each subpackage exposes a `New(...)` constructor returning a `*Model` that implements `form.Field`. Each text-valued field SHALL accept an optional validator function via an `Option`. The `savefield` SHALL be a valueless focus placeholder: it carries no value, always validates, and does NOT claim Enter — submission when it holds focus is provided by the form-level "Enter on the last visible field submits" rule rather than by the field itself. The `selectfield` SHALL claim Enter only while its list is in the filtering state (so Enter accepts the filter); when not filtering it SHALL NOT claim Enter, so a terminal selectfield submits via the same form-level last-field rule. The toolkit SHALL NOT provide a `WithSubmitOnEnter` selectfield option.

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

#### Scenario: savefield does not claim Enter
- **WHEN** a `savefield` is focused inside a form
- **THEN** its `Keys()` do not claim Enter
- **AND** pressing Enter while it holds focus submits the form via the form-level last-field rule

#### Scenario: selectfield claims Enter only while filtering
- **WHEN** a focused `selectfield` is not filtering
- **THEN** its `Keys()` do not claim Enter, so a terminal selectfield submits via the form's last-field rule
- **AND WHEN** the list is in the filtering state
- **THEN** its `Keys()` claim Enter so the gesture accepts the filter instead of submitting the form
