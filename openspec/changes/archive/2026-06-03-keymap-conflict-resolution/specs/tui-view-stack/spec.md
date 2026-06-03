## MODIFIED Requirements

### Requirement: Overlay wrapper with esc-to-dismiss
The system SHALL provide a `screen.Overlay(parent, child)` function that wraps a child Screen with a parent reference. The overlay SHALL handle esc by sending `screen.Dismiss()` when the inner screen is not capturing input, and SHALL forward esc to the inner screen when it is capturing input (free-text), as gated by `CapturingInput`. The esc binding SHALL live on a `screen.KeyMap` value carried by the overlay and SHALL be exposed as a `keymap.Chord` from the overlay's `Chords()`.

The `Screen` interface SHALL embed `keymap.Map` (`Chords() []keymap.Group`). An overlay SHALL produce its `Chords()` by concatenating the inner screen's `Chords()` (the inner's full subtree) ahead of its own esc chord, so the overlay contributes a single aggregated group list highest-priority first. Conflict resolution — performed by the `keymap` package — SHALL subtract the overlay's esc from help when the inner subtree already claims esc, replacing the previous bespoke `hasEsc` dedup. The overlay SHALL NOT hand-compose or de-duplicate help itself.

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
- **THEN** the overlay's own esc chord is absent from help (no duplicate esc)

#### Scenario: Overlay esc shown when inner does not claim it
- **WHEN** the overlay's resolved help is read
- **AND** no inner-subtree binding claims esc
- **THEN** the overlay's esc chord appears in help

### Requirement: InputCapturer delegation through overlays
The overlay wrapper SHALL delegate InputCapturer to its inner screen. This ensures that app.Model and the overlay suppress their own keybindings (? and esc respectively) when the inner screen is capturing text input. `CapturingInput` remains the mechanism for free-text capture and is independent of `keymap` conflict resolution.

#### Scenario: Overlay propagates capturing state
- **WHEN** the inner screen implements InputCapturer and returns true
- **THEN** the overlay's CapturingInput() SHALL return true

#### Scenario: App suppresses help toggle during capture
- **WHEN** the user presses ?
- **AND** CapturingInput(m.active) returns true
- **THEN** app SHALL forward ? to the active screen instead of toggling help
