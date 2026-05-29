## ADDED Requirements

### Requirement: Centralized error display in help bar row
`tui.Model` SHALL intercept `error` messages returned by screens. When an error is active, the help bar row SHALL display the error message in red/bold instead of the normal help bindings.

#### Scenario: Error replaces help bar
- **WHEN** a screen returns an `error` message
- **THEN** the help bar row SHALL display the error message in red/bold
- **AND** the normal help bindings SHALL be hidden

#### Scenario: No error shows normal help
- **WHEN** no error is active
- **THEN** the help bar row SHALL display the normal help bindings

### Requirement: Any keypress clears the error
When an error is active, any keypress SHALL clear the error bar and restore the normal help display. The keypress SHALL NOT be consumed — it SHALL be forwarded to the active screen for normal processing.

#### Scenario: Esc clears error with no side effect
- **WHEN** an error is active and the user presses esc
- **THEN** the error SHALL be cleared
- **AND** the esc SHALL be forwarded to the active screen

#### Scenario: Action key clears error and acts
- **WHEN** an error is active and the user presses space
- **THEN** the error SHALL be cleared
- **AND** the space SHALL be forwarded to the active screen for normal processing

### Requirement: Errors are replaced by newer errors
When a new `error` message arrives while an existing error is active, the new error SHALL replace the old one.

#### Scenario: Error replacement
- **WHEN** an error is active and a new error message arrives
- **THEN** the displayed error SHALL be the new error message

### Requirement: Screens stop rendering errors locally
Screens that currently handle `case error:` for ambient display (projectlist, projectpicker) SHALL remove their local error fields and error rendering. They SHALL allow `error` messages to propagate to `tui.Model`. Save-error overlays (taskedit, projectedit, projectstatus, taskstatus) SHALL keep internal error state for blocking form re-fire, but SHALL return the error as a cmd yielding an `error` message for the app error bar instead of rendering it locally.

#### Scenario: Projectlist error goes to app
- **WHEN** a project list load fails
- **THEN** the error SHALL be displayed in the app error bar
- **AND** the project list SHALL NOT render the error itself

#### Scenario: Save error in overlay goes to app
- **WHEN** a save fails in taskedit
- **THEN** the overlay SHALL block form re-fire internally
- **AND** return a cmd yielding an `error` for the app error bar
- **AND** the overlay SHALL NOT render the error message itself

#### Scenario: Esc clears both overlay block and error bar
- **WHEN** an overlay is in error-blocked state and the user presses esc
- **THEN** the app SHALL clear the error bar (any keypress clears it)
- **AND** the esc SHALL reach the overlay, which clears its block state and resumes the form
