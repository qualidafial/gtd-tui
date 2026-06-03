## MODIFIED Requirements

### Requirement: App KeyMap and help composition
The root model SHALL expose an exported `KeyMap` field of type `tui.KeyMap` defined in `tui/keymap.go`. The KeyMap SHALL carry the app-wide bindings (`Quit`, `Help`) and SHALL expose them as `keymap.Group`s via `Chords()`. The root model SHALL produce its own `Chords()` by concatenating the active screen's `Chords()` (the active subtree) ahead of the app KeyMap's groups, so the active screen's bindings take priority over the app's. The root model's `ShortHelp`/`FullHelp` SHALL be derived from `keymap.Resolve(render, m.Chords()...)`: the short bar keeps `Vis == Short` chords flattened in priority order; full help keeps `Vis ∈ {Short, Full}` chords as group rows. The `Help` binding SHALL be disabled (and thus claim/display nothing) while the active screen is capturing input, so the `?` keystroke reaches the screen.

#### Scenario: App help merges with active-screen help via resolution
- **WHEN** `Model.ShortHelp` is invoked
- **THEN** the returned bindings SHALL be the `keymap.Resolve` projection of the active screen's chords followed by the app KeyMap's chords
- **AND** a key claimed by the active screen SHALL NOT also appear under an app binding

#### Scenario: Help binding suppressed during input capture
- **WHEN** the active screen reports `CapturingInput() == true`
- **THEN** the `Help` binding in the app KeyMap SHALL be disabled
- **AND** the `?` keystroke SHALL be forwarded to the active screen instead of toggling help

#### Scenario: Active screen binding wins a conflict with the app
- **WHEN** the active subtree claims a key the app KeyMap also binds
- **THEN** the keypress routes to the active screen via `Handles`
- **AND** the app binding is subtracted from the resolved help

### Requirement: Per-component KeyMap files
Each TUI component or page that owns key bindings SHALL declare them in a sibling `keymap.go` file as an exported `KeyMap` struct with a `DefaultKeyMap()` constructor and a `Chords() []keymap.Group` method (replacing the former `ShortHelp`/`FullHelp` pair). The owning Model SHALL store the KeyMap as an exported field (`Model.KeyMap`), not as a private variable, so tests and parent screens can reach the binding objects without re-declaring types.

Affected components: `tui` (app), `tui/components/screen` (overlay), `tui/components/tabcontainer`, `tui/components/querybar`.

Affected pages: `tui/pages/inbox`, `tui/pages/projects` (project list), `tui/pages/projects/projectview`, `tui/pages/tasks/tasklist`.

Form-based pages (`projectedit`, `projectpicker`, `projectstatus`, `taskedit`, `taskstatus`, `itemcapture`, `clarify`) SHALL NOT declare a separate `KeyMap` struct; instead they SHALL implement `Chords()` by aggregating `form.Chords()` (the focused field's and form's bindings) with their own esc/back binding.

#### Scenario: Page keymap is reachable from tests
- **WHEN** a test constructs a `tasklist.Model` and reads `m.KeyMap.MoveUp.Enabled()`
- **THEN** the test SHALL observe the same `key.Binding` instance the running model uses

#### Scenario: Form-based page advertises form bindings
- **WHEN** a form-based page's `Chords()` is invoked
- **THEN** the returned groups SHALL be `form.Chords()` followed by the page's own esc/back chord
