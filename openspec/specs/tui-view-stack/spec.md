# tui-view-stack Specification

## Purpose
Defines the view-stack primitives: push and dismiss commands, the esc-to-dismiss overlay wrapper, the Popper pop interface, InputCapturer delegation through overlays, and the init-driven data lifecycle.

## Requirements

### Requirement: Push command for view transitions
The system SHALL provide a `screen.Push(child)` command that produces a `PushMsg`. app.Model SHALL handle PushMsg by wrapping the current active screen in an overlay with the child as the inner screen.

#### Scenario: Tab screen pushes a child view
- **WHEN** a screen inside tabContainer returns `screen.Push(child)` from Update
- **THEN** app.Model SHALL wrap the current active (tabContainer) as the overlay parent
- **AND** the child screen SHALL be the overlay's inner screen
- **AND** app SHALL call Init() on the new overlay

#### Scenario: Overlaid screen pushes deeper
- **WHEN** an overlaid screen returns `screen.Push(child)` from Update
- **THEN** app.Model SHALL wrap the current overlay as the parent of a new overlay
- **AND** the child screen SHALL be the new overlay's inner screen

### Requirement: Dismiss command for view dismissal
The system SHALL provide a `screen.Dismiss()` command that produces a `DismissMsg`. app.Model SHALL handle DismissMsg by popping the active screen via the Popper interface.

#### Scenario: Inner screen dismisses after save
- **WHEN** an inner screen returns `m, screen.Dismiss()` from Update
- **THEN** app.Model SHALL pop the overlay and restore the parent as active
- **AND** app SHALL call Init() on the parent

#### Scenario: Dismiss at root is a no-op
- **WHEN** app.Model receives DismissMsg and the active screen does not satisfy Popper
- **THEN** nothing happens

#### Scenario: Inner screen returns valid model on dismiss
- **WHEN** an inner screen sends Dismiss
- **THEN** the inner screen SHALL return its own model (not nil) from Update

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

### Requirement: Popper interface for overlay pop
The system SHALL define a `Popper` interface with a `Pop() Screen` method. The overlay wrapper SHALL satisfy Popper by returning its parent screen.

#### Scenario: Overlay satisfies Popper
- **WHEN** app.Model receives DismissMsg
- **AND** the active screen satisfies Popper
- **THEN** app SHALL call Pop() and set the result as the new active screen

### Requirement: InputCapturer delegation through overlays
The overlay wrapper SHALL delegate InputCapturer to its inner screen. This ensures that app.Model and the overlay suppress their own keybindings (? and esc respectively) when the inner screen is capturing text input. `CapturingInput` remains the mechanism for free-text capture and is independent of `keymap` conflict resolution.

#### Scenario: Overlay propagates capturing state
- **WHEN** the inner screen implements InputCapturer and returns true
- **THEN** the overlay's CapturingInput() SHALL return true

#### Scenario: App suppresses help toggle during capture
- **WHEN** the user presses ?
- **AND** CapturingInput(m.active) returns true
- **THEN** app SHALL forward ? to the active screen instead of toggling help

### Requirement: Init-driven data lifecycle
The system SHALL use `InitMsg` / `InitCmd` to trigger screen initialization on view transitions. Every screen SHALL reload its own data in Init().

#### Scenario: Push triggers child Init
- **WHEN** a child view is pushed via PushMsg
- **THEN** app SHALL call Init() on the new active screen

#### Scenario: Pop triggers parent Init
- **WHEN** an overlay pops via DismissMsg
- **THEN** app SHALL call Init() on the restored parent screen
- **AND** the parent SHALL reload its data from the database

#### Scenario: Cross-view coordination via init
- **WHEN** a screen modifies data and dismisses
- **THEN** the parent screen SHALL reload fresh data on Init()
- **AND** no explicit coordination messages (e.g. TasksChangedMsg) are needed