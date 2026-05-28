## MODIFIED Requirements

### Requirement: Quick-create project with "+" or "insert" key
The system SHALL allow creating a new project by pressing "+" or "insert", which pushes the project edit overlay with an empty project. On submit, it creates an open project with the entered fields.

#### Scenario: Create new project
- **WHEN** user presses "+" or "insert"
- **THEN** the project edit overlay SHALL be pushed with an empty project
- **WHEN** user fills in fields and submits
- **THEN** the system SHALL call CreateProject with status=open and the entered fields
- **AND** the editor overlay SHALL dismiss
- **AND** the project view screen SHALL be pushed for the newly created project

#### Scenario: Cancel project creation
- **WHEN** user presses "+" or "insert" to open the editor
- **AND** user presses escape
- **THEN** the overlay SHALL be dismissed without creating a project

## ADDED Requirements

### Requirement: Edit project from project list with "e" key
The system SHALL allow editing the selected project by pressing "e", which pushes the project edit overlay with the selected project. The "e" key SHALL be enabled when a project is selected and disabled when the list is empty.

#### Scenario: Edit selected project
- **WHEN** user selects a project
- **AND** presses "e"
- **THEN** the project edit overlay SHALL be pushed with the selected project

#### Scenario: Edit disabled when no selection
- **WHEN** the list is empty
- **THEN** the "e" keybinding SHALL be disabled

#### Scenario: Return from edit
- **WHEN** user saves or cancels the edit
- **THEN** the overlay SHALL dismiss
- **AND** the project list SHALL reinitialize
