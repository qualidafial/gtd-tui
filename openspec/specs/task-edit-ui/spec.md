# task-edit-ui Specification

## Purpose
Defines the task editor overlay: form fields, defaults, read-only header, save/cancel behavior, and error handling.

## Requirements

### Requirement: Task editor form fields
The task editor SHALL present a form for the editable fields of a task: Title, Description, Assignee, Due, and Defer Until, in that order. Title SHALL be required (a non-empty validation error blocks saving). Due and Defer Until SHALL use the shared date field (natural-language and `YYYY-MM-DD[ HH:MM]` input). Status SHALL NOT be editable from the editor; status changes are made through the transition confirmation overlay. The Kind field SHALL NOT appear.

#### Scenario: Editor opens on an existing task
- **WHEN** the editor opens for an existing task
- **THEN** the form presents Title, Description, Assignee, Due, and Defer Until

#### Scenario: Title is required
- **WHEN** the user clears the Title and attempts to submit
- **THEN** a validation error is shown and the task is not saved

### Requirement: Default values for new tasks
When the editor is opened for a new task (no ID assigned), Status SHALL default to open.

#### Scenario: New task defaults
- **WHEN** the editor opens for a new task
- **THEN** Status is open

### Requirement: Non-editable properties header
When editing an existing task (an ID is assigned), the editor SHALL render a compact, dimmed header above the form showing read-only properties: Task ID, Created, Updated, and Status. Each property SHALL be a label/value line. Created and Updated SHALL render their timestamps in the local timezone. The header SHALL NOT appear when creating a new task.

#### Scenario: Header shown for an existing task
- **WHEN** the editor is opened for an existing task
- **THEN** a dimmed header shows Task ID, Created, Updated, and Status above the form

#### Scenario: No header for a new task
- **WHEN** the editor is opened to create a task
- **THEN** no read-only properties header is shown

### Requirement: Status property with relative change time
The Status line in the read-only header SHALL show the task's status name (first letter capitalized) followed by a relative WHEN of the last status change, in parentheses, on the same line: `Status: <Status> (<WHEN>)`. The WHEN SHALL be computed from StatusChangedAt against the current time using the shared relative-time formatter's past ladder (`today`, `Nd` up to 30 days, then an absolute `YYYY-MM-DD` date; a change earlier the same day renders the clock time).

#### Scenario: Open task changed three days ago
- **WHEN** an existing open task's StatusChangedAt was 3 days ago
- **THEN** the header shows `Status: Open (3d)`

#### Scenario: Done task changed today
- **WHEN** an existing task is done and its StatusChangedAt is today (date-only)
- **THEN** the header shows `Status: Done (today)`

### Requirement: Save creates or updates
Submitting the form SHALL create the task when it has no ID and update it otherwise. On success the editor SHALL dismiss its overlay and broadcast that tasks have changed so open task lists refresh.

#### Scenario: Create on submit
- **WHEN** the form is submitted for a task with no ID
- **THEN** the task is created
- **AND** the overlay is dismissed and lists are refreshed

#### Scenario: Update on submit
- **WHEN** the form is submitted for a task with an ID
- **THEN** the task is updated
- **AND** the overlay is dismissed and lists are refreshed

### Requirement: Save-error surfacing and retry
When a create or update fails, the editor SHALL render the error in red beneath the form rather than silently dropping it. Pressing esc SHALL clear the error and return the form to its editable state so the user can adjust input and retry; pressing esc again SHALL back out. While an error is showing, other keys SHALL be ignored so the form's completed state cannot re-fire the save.

#### Scenario: Save failure is shown
- **WHEN** a save fails
- **THEN** the error message is displayed beneath the form

#### Scenario: Esc clears the error for retry
- **WHEN** an error is showing and the user presses esc
- **THEN** the error is cleared and the form returns to its editable state

#### Scenario: Other keys ignored during an error
- **WHEN** an error is showing and the user presses a non-esc key
- **THEN** the keypress is ignored and no save is re-fired

### Requirement: Back out without saving
Pressing esc while editing (with no error showing) SHALL abort the form and dismiss the overlay without saving changes.

#### Scenario: Esc backs out
- **WHEN** the user presses esc while editing
- **THEN** the overlay is dismissed and no changes are saved