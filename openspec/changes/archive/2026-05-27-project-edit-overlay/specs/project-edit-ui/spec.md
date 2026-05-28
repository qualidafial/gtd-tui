## ADDED Requirements

### Requirement: Project editor form fields
The project editor SHALL present a huh form for the editable fields of a project: Title, Outcome, Description, and Due, in that order. Title and Outcome SHALL be required (non-empty validation blocks saving). Description SHALL be a multi-line text field (optional). Due SHALL use the shared date field (natural-language and `YYYY-MM-DD[ HH:MM]` input). Status SHALL NOT be editable from the editor; status transitions remain in the dedicated confirmation overlays.

#### Scenario: Editor opens on an existing project
- **WHEN** the editor opens for an existing project
- **THEN** the form presents Title, Outcome, Description, and Due pre-populated with the project's current values

#### Scenario: Editor opens for a new project
- **WHEN** the editor opens for a new project
- **THEN** the form presents Title, Outcome, Description, and Due with empty/nil values

#### Scenario: Title is required
- **WHEN** the user clears the Title and attempts to submit
- **THEN** a validation error is shown and the project is not saved

#### Scenario: Outcome is required
- **WHEN** the user clears the Outcome and attempts to submit
- **THEN** a validation error is shown and the project is not saved

### Requirement: Read-only header for existing projects
When editing an existing project (ID != 0), the editor SHALL render a compact, dimmed header above the form showing read-only properties: Project ID, Status (with relative time from StatusChangedAt), Created, and Updated. Each property SHALL be a label/value line. Created and Updated SHALL render their timestamps in the local timezone. The header SHALL NOT appear when creating a new project.

#### Scenario: Header shown for existing project
- **WHEN** the editor is opened for an existing project
- **THEN** a dimmed header shows Project ID, Status, Created, and Updated above the form

#### Scenario: No header for new project
- **WHEN** the editor is opened to create a project
- **THEN** no read-only properties header is shown

### Requirement: Status property with relative change time
The Status line in the read-only header SHALL show the project's status name (first letter capitalized) followed by a relative time of the last status change, in parentheses: `Status: <Status> (<WHEN>)`. The WHEN SHALL be computed from StatusChangedAt against the current time using the shared relative-time formatter.

#### Scenario: Open project changed three days ago
- **WHEN** an existing open project's StatusChangedAt was 3 days ago
- **THEN** the header shows `Status: Open (3d)`

#### Scenario: Done project changed today
- **WHEN** an existing project is done and its StatusChangedAt is today
- **THEN** the header shows `Status: Done (today)`

### Requirement: Default values for new projects
When the editor is opened for a new project (no ID assigned), Status SHALL default to open.

#### Scenario: New project defaults
- **WHEN** the editor opens for a new project
- **THEN** Status is open

### Requirement: Save creates or updates
Submitting the form SHALL create the project when it has no ID and update it otherwise. On success the editor SHALL dismiss its overlay. For updates, dismiss is sufficient (the parent reloads on dismiss). For creates, the editor SHALL dismiss and then push the project view screen for the newly created project, so the user can immediately add tasks. The dismiss-then-push SHALL be sequenced using `tea.Sequence` to guarantee ordering (dismiss completes before push). For testability, the sequence cmd can be cast to `[]tea.Cmd` to inspect the individual commands.

#### Scenario: Create on submit
- **WHEN** the form is submitted for a project with no ID
- **THEN** the project is created via CreateProject
- **AND** the overlay is dismissed
- **AND** the project view screen is pushed for the new project

#### Scenario: Update on submit
- **WHEN** the form is submitted for a project with an ID
- **THEN** the project is updated via UpdateProject
- **AND** the overlay is dismissed

### Requirement: Save-error surfacing and retry
When a create or update fails, the editor SHALL render the error in red beneath the form. Pressing esc SHALL clear the error and return the form to its editable state so the user can adjust input and retry; pressing esc again SHALL dismiss the overlay. While an error is showing, other keys SHALL be ignored so the form's completed state cannot re-fire the save.

#### Scenario: Save failure is shown
- **WHEN** a save fails
- **THEN** the error message is displayed beneath the form

#### Scenario: Esc clears the error for retry
- **WHEN** an error is showing and the user presses esc
- **THEN** the error is cleared and the form returns to its editable state

#### Scenario: Other keys ignored during error
- **WHEN** an error is showing and the user presses a non-esc key
- **THEN** the keypress is ignored and no save is re-fired

### Requirement: Back out without saving
Pressing esc while editing (with no error showing) SHALL abort the form and dismiss the overlay without saving changes.

#### Scenario: Esc backs out
- **WHEN** the user presses esc while editing
- **THEN** the overlay is dismissed and no changes are saved