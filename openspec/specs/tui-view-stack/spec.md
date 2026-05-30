## ADDED Requirements

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

### Requirement: Popper interface for overlay pop
The system SHALL define a `Popper` interface with a `Pop() Screen` method. The overlay wrapper SHALL satisfy Popper by returning its parent screen.

#### Scenario: Overlay satisfies Popper
- **WHEN** app.Model receives DismissMsg
- **AND** the active screen satisfies Popper
- **THEN** app SHALL call Pop() and set the result as the new active screen

### Requirement: InputCapturer delegation through overlays
The overlay wrapper SHALL delegate InputCapturer to its inner screen. This ensures that app.Model and the overlay suppress their own keybindings (? and esc respectively) when the inner screen is capturing text input.

#### Scenario: Overlay propagates capturing state
- **WHEN** the inner screen implements InputCapturer and returns true
- **THEN** the overlay's CapturingInput() SHALL return true

#### Scenario: App suppresses help toggle during capture
- **WHEN** app.Model receives ? key press
- **AND** CapturingInput(m.active) returns true
- **THEN** app SHALL forward ? to the active screen instead of toggling help

#### Scenario: huh form receives all keys while editing
- **WHEN** a huh form screen is focused (CapturingInput = true)
- **AND** the user presses esc
- **THEN** the overlay suppresses its own esc
- **AND** the form receives esc and aborts
- **AND** the screen sends Dismiss()

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