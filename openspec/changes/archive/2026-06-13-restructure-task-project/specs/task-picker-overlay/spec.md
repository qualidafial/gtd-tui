## ADDED Requirements

### Requirement: Picker shows standalone open tasks
The task picker overlay SHALL display a `selectfield` (the in-house form component) populated with standalone open tasks loaded from TaskService — tasks whose ProjectID is nil and whose status is open. The select SHALL be sized to the overlay's available height. There is no `(none)` option: the overlay's purpose is to choose a task. The calling screen is responsible for only opening the picker when at least one candidate exists, so the picker does not need to render an empty state.

#### Scenario: Standalone open tasks listed
- **WHEN** the picker overlay is opened
- **THEN** all standalone open tasks SHALL be listed as options
- **AND** tasks that already belong to a project SHALL NOT appear

### Requirement: Picker broadcasts the selection on confirm
The task picker SHALL be selection-only: on confirm it SHALL emit a selection message carrying the chosen task to the calling screen and dismiss, without applying any change itself. It SHALL NOT call any link, update, or other mutating operation. Applying the selection (e.g. linking the task into a project) is the calling screen's responsibility. The selection message SHALL be delivered to the parent after the overlay is dismissed (via `screen.Dismiss` with the emit sequenced after the pop) so the parent is the active screen when it arrives.

#### Scenario: Selection broadcast on confirm
- **WHEN** the user selects a task and confirms
- **THEN** the picker SHALL emit a selection message carrying the chosen task
- **AND** the picker SHALL dismiss
- **AND** the picker SHALL NOT call any mutating service operation

#### Scenario: Parent receives the selection after dismissal
- **WHEN** the picker confirms a selection
- **THEN** the selection message SHALL be delivered to the calling screen once it is active again

### Requirement: Picker dismisses on cancel
Pressing esc SHALL dismiss the picker without emitting a selection.

#### Scenario: Cancel without selecting
- **WHEN** the user presses esc
- **THEN** the overlay SHALL dismiss
- **AND** no selection message SHALL be emitted

### Requirement: Picker does not own apply errors
Because the picker does not apply the selection, it SHALL NOT display apply/link errors. Error handling for the applied change is owned by the calling screen. The picker's own responsibilities are limited to loading candidates, rendering the select, and emitting the selection.

#### Scenario: Apply error handled by caller, not picker
- **WHEN** applying the selection fails after the picker has dismissed
- **THEN** the error SHALL be surfaced by the calling screen
- **AND** the picker SHALL already be dismissed and SHALL NOT display the error
