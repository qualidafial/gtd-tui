## ADDED Requirements

### Requirement: App KeyMap and help composition
The root model SHALL expose an exported `KeyMap` field of type `tui.KeyMap` defined in `tui/keymap.go`. The KeyMap SHALL carry the app-wide bindings (`Quit`, `Help`) and SHALL implement `help.KeyMap` (`ShortHelp`/`FullHelp`). The root model's own `ShortHelp`/`FullHelp` SHALL compose the app KeyMap with the active screen's help via `slices.Concat`. The `Help` binding SHALL be disabled while the active screen is capturing input so the keystroke reaches the screen.

#### Scenario: App help merges with active-screen help
- **WHEN** `Model.ShortHelp` is invoked
- **THEN** the returned bindings SHALL be the app KeyMap's `ShortHelp` concatenated with the active screen's `ShortHelp`

#### Scenario: Help binding suppressed during input capture
- **WHEN** the active screen reports `CapturingInput() == true`
- **THEN** the `Help` binding in the app KeyMap SHALL be disabled
- **AND** the `?` keystroke SHALL be forwarded to the active screen instead of toggling help

### Requirement: Per-component KeyMap files
Each TUI component or page that owns key bindings SHALL declare them in a sibling `keymap.go` file as an exported `KeyMap` struct with a `DefaultKeyMap()` constructor and `ShortHelp`/`FullHelp` methods. The owning Model SHALL store the KeyMap as an exported field (`Model.KeyMap`), not as a private variable, so tests and parent screens can reach the binding objects without re-declaring types.

Affected components: `tui` (app), `tui/components/screen` (overlay), `tui/components/tabcontainer`, `tui/components/querybar`.

Affected pages: `tui/pages/projects` (project list), `tui/pages/projects/projectview`, `tui/pages/tasks/tasklist`.

Huh-form pages (`projectedit`, `projectpicker`, `projectstatus`, `taskedit`, `taskstatus`) SHALL NOT declare a separate `KeyMap` struct; instead they SHALL implement `ShortHelp`/`FullHelp` by forwarding `form.KeyBinds()`.

#### Scenario: Page keymap is reachable from tests
- **WHEN** a test constructs a `tasklist.Model` and reads `m.KeyMap.MoveUp.Enabled()`
- **THEN** the test SHALL observe the same `key.Binding` instance the running model uses

#### Scenario: Huh-form page advertises form bindings
- **WHEN** a huh-form page's `ShortHelp` is invoked
- **THEN** the returned bindings SHALL be `form.KeyBinds()` from the underlying huh form

## MODIFIED Requirements

### Requirement: Quit command exits application
The system SHALL respond to the quit key binding (Ctrl+C) by terminating the application cleanly. The `?` key SHALL toggle help display, suppressed when the active screen is capturing input. Both bindings SHALL be declared on the root model's `KeyMap` field in `tui/keymap.go`.

#### Scenario: Quit with Ctrl+C
- **WHEN** user presses Ctrl+C
- **THEN** application sends tea.Quit command
- **AND** program exits cleanly

#### Scenario: Toggle help with ?
- **WHEN** user presses ?
- **AND** the active screen is not capturing input
- **THEN** help display is toggled

#### Scenario: ? suppressed during input capture
- **WHEN** user presses ?
- **AND** CapturingInput(active) returns true
- **THEN** ? is forwarded to the active screen