## ADDED Requirements

### Requirement: Picker shows open projects and none option
The project picker overlay SHALL display a huh.Select with all open projects loaded from ProjectService, plus a "(none)" option to unlink the task from any project. The select SHALL be sized to the overlay's available height.

#### Scenario: Open projects listed
- **WHEN** the picker overlay is opened
- **THEN** all open projects SHALL be listed as options
- **AND** a "(none)" option SHALL appear to allow unlinking

#### Scenario: Current project pre-selected
- **WHEN** the picker is opened for a task that already has a ProjectID
- **THEN** the matching project SHALL be pre-selected in the list

#### Scenario: No project pre-selects none
- **WHEN** the picker is opened for a task with nil ProjectID
- **THEN** "(none)" SHALL be pre-selected

### Requirement: Picker updates task on confirm
On confirm, the picker SHALL update the task's ProjectID via TaskService.UpdateTask and dismiss.

#### Scenario: Assign project
- **WHEN** the user selects a project and confirms
- **THEN** the task's ProjectID SHALL be set to the selected project's ID
- **AND** TaskService.UpdateTask SHALL be called
- **AND** the overlay SHALL dismiss

#### Scenario: Unlink project
- **WHEN** the user selects "(none)" and confirms
- **THEN** the task's ProjectID SHALL be set to nil
- **AND** TaskService.UpdateTask SHALL be called
- **AND** the overlay SHALL dismiss

#### Scenario: No change skips update
- **WHEN** the user confirms the same project that was already assigned
- **THEN** no update SHALL be performed
- **AND** the overlay SHALL dismiss

### Requirement: Picker dismisses on cancel
Pressing esc SHALL dismiss the picker without modifying the task.

#### Scenario: Cancel without change
- **WHEN** the user presses esc
- **THEN** the overlay SHALL dismiss
- **AND** the task SHALL not be modified

### Requirement: Picker is self-contained
The picker SHALL load projects, perform the update, and dismiss without requiring the parent screen to handle any result messages. Errors from UpdateTask SHALL be displayed within the picker overlay.

#### Scenario: Save error displayed
- **WHEN** UpdateTask fails
- **THEN** the error SHALL be displayed in the picker
- **AND** the overlay SHALL remain open
