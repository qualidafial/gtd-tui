## MODIFIED Requirements

### Requirement: Quick-create project with "+" or "insert" key
The system SHALL allow creating a new project by pressing "+" or "insert", which pushes a title-only input overlay. On submit, it creates an open project with the entered title.

#### Scenario: Create new project
- **WHEN** user presses "+" or "insert"
- **THEN** a title input overlay SHALL be pushed
- **WHEN** user enters a title and presses enter
- **THEN** the system SHALL call CreateProject with status=open and the entered title
- **AND** dismiss the overlay (triggering reload)

#### Scenario: Cancel project creation
- **WHEN** user presses "+" or "insert" to open the create overlay
- **AND** user presses escape
- **THEN** the overlay SHALL be dismissed without creating a project

#### Scenario: Empty title rejected
- **WHEN** user presses enter with an empty title
- **THEN** the system SHALL NOT create a project
- **AND** the overlay SHALL remain open

### Requirement: Enter project view with enter key
The system SHALL allow entering a project's detail view by pressing enter on the selected project, which pushes the project view screen as an overlay.

#### Scenario: Enter project view
- **WHEN** user selects a project
- **AND** presses enter
- **THEN** the project view screen SHALL be pushed as an overlay
- **AND** the project view SHALL display the selected project's details and tasks

#### Scenario: Return from project view
- **WHEN** user presses esc in the project view
- **THEN** the project view SHALL dismiss
- **AND** the project list SHALL reinitialize

### Requirement: Keybindings reflect selected project state
The system SHALL enable/disable action keybindings based on the selected project's status, and update help text accordingly.

#### Scenario: Open project selected
- **WHEN** an open project is selected
- **THEN** space label SHALL be "complete", delete SHALL be enabled, "s" SHALL be enabled, reorder SHALL be conditionally enabled

#### Scenario: Someday project selected
- **WHEN** a someday project is selected
- **THEN** space label SHALL be "reopen", delete SHALL be enabled, "s" SHALL be disabled, reorder SHALL be conditionally enabled

#### Scenario: Done project selected
- **WHEN** a done project is selected
- **THEN** space label SHALL be "reopen", delete SHALL be disabled, "s" SHALL be disabled, reorder SHALL be disabled

#### Scenario: No project selected
- **WHEN** the list is empty
- **THEN** all action bindings except "+" (new) SHALL be disabled