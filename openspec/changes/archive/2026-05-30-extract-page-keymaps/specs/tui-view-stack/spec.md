## MODIFIED Requirements

### Requirement: Overlay wrapper with esc-to-dismiss
The system SHALL provide a `screen.Overlay(parent, child)` function that wraps a child Screen with a parent reference. The overlay SHALL handle esc by sending `screen.Dismiss()` when the inner screen is not capturing input. The esc binding SHALL live on a `screen.KeyMap` value carried by the overlay; the overlay SHALL contribute that binding to its own `ShortHelp`/`FullHelp` (the `Screen` interface SHALL embed `help.KeyMap` directly, so overlays and their children compose help by `slices.Concat` rather than via a separate `KeyMap() help.KeyMap` method).

#### Scenario: Esc pops overlay when not capturing
- **WHEN** the user presses esc
- **AND** the inner screen is not capturing input
- **THEN** the overlay SHALL send Dismiss()

#### Scenario: Esc forwarded when capturing input
- **WHEN** the user presses esc
- **AND** the inner screen is capturing input (e.g. huh form editing)
- **THEN** the overlay SHALL forward esc to the inner screen
- **AND** the overlay SHALL NOT send Dismiss()

#### Scenario: Overlay help includes esc
- **WHEN** the overlay's `ShortHelp` (or `FullHelp`) is read
- **THEN** it SHALL include the `Back` binding from the overlay's KeyMap
- **AND** it SHALL include all bindings from the inner screen's help
- **AND** if the inner screen already advertises an esc binding, the overlay SHALL NOT add a second one