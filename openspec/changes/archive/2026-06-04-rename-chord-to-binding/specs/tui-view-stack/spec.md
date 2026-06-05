## MODIFIED Requirements

### Requirement: Overlay wrapper with esc-to-dismiss
The system SHALL provide a `screen.Overlay(parent, child)` function that wraps a child Screen with a parent reference. The overlay SHALL handle esc by sending `screen.Dismiss()` when the inner screen is not capturing input, and SHALL forward esc to the inner screen when it is capturing input (free-text), as gated by `CapturingInput`. The esc binding SHALL live on a `screen.KeyMap` value carried by the overlay and SHALL be exposed as a `keymap.Binding` from the overlay's `Keys()`.

The `Screen` interface SHALL embed `keymap.Map` (`Keys() []keymap.Group`). An overlay SHALL produce its `Keys()` by concatenating the inner screen's `Keys()` (the inner's full subtree) ahead of its own esc binding, so the overlay contributes a single aggregated group list highest-priority first. Conflict resolution — performed by the `keymap` package — SHALL subtract the overlay's esc from help when the inner subtree already claims esc, replacing the previous bespoke `hasEsc` dedup. The overlay SHALL NOT hand-compose or de-duplicate help itself.

#### Scenario: Esc pops overlay when not capturing
- **WHEN** the user presses esc
- **AND** the inner screen is not capturing input
- **THEN** the overlay sends `screen.Dismiss()`

#### Scenario: Esc forwarded when capturing input
- **WHEN** the user presses esc
- **AND** the inner screen is capturing input (free-text entry)
- **THEN** the overlay SHALL forward esc to the inner screen

#### Scenario: Overlay esc subtracted from help when inner claims it
- **WHEN** the overlay's resolved help is read
- **AND** the inner subtree already claims an esc binding
- **THEN** the overlay's own esc binding is absent from help (no duplicate esc)

#### Scenario: Overlay esc shown when inner does not claim it
- **WHEN** the overlay's resolved help is read
- **AND** no inner-subtree binding claims esc
- **THEN** the overlay's esc binding appears in help
