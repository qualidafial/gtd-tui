# project-list-ui Specification

## Purpose
Defines the interactive behavior of the project list screen, including loading, navigation, keybindings, and project state transitions available from the list.

### Requirement: Project list loads projects on init
The system SHALL load all projects from ProjectService on initialization and display them in the list sorted by status group (open, someday, then done/dropped) and order_key within each group.

#### Scenario: Initial load
- **WHEN** the project list screen initializes
- **THEN** it SHALL issue a command to load all projects via ListProjects with no filter
- **AND** display the results in the list

#### Scenario: Reload on init after overlay dismiss
- **WHEN** an overlay is dismissed and the project list receives Init
- **THEN** it SHALL reload the project list from the service

### Requirement: Project list loads task counts
The system SHALL load task counts (pending and total) for all displayed projects after the project list loads.

#### Scenario: Task counts loaded
- **WHEN** projects are loaded
- **THEN** the system SHALL fetch task counts for all loaded project IDs in a single batch query
- **AND** associate counts with each project row for rendering

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

### Requirement: Toggle project status with space key
The system SHALL allow toggling project status with space: completing open projects (with confirmation), and reopening someday/done/dropped projects (immediately).

#### Scenario: Complete open project
- **WHEN** user selects an open project
- **AND** presses space
- **THEN** a confirmation overlay SHALL be pushed
- **WHEN** user confirms
- **THEN** CompleteProject SHALL be called with cascade=true
- **AND** the overlay SHALL dismiss

#### Scenario: Reopen someday project with space
- **WHEN** user selects a someday project
- **AND** presses space
- **THEN** ReopenProject SHALL be called immediately (no confirmation)
- **AND** the list SHALL reload

#### Scenario: Reopen done/dropped project with space
- **WHEN** user selects a done or dropped project
- **AND** presses space
- **THEN** ReopenProject SHALL be called immediately (no confirmation)
- **AND** the list SHALL reload

### Requirement: Drop project with delete key
The system SHALL allow dropping an open or someday project by pressing delete/backspace, which pushes a confirmation overlay.

#### Scenario: Drop project
- **WHEN** user selects an open or someday project
- **AND** presses delete
- **THEN** a confirmation overlay SHALL be pushed showing task cascade info
- **WHEN** user confirms
- **THEN** DropProject SHALL be called with cascade=true
- **AND** the overlay SHALL dismiss

#### Scenario: Drop disabled for done/dropped
- **WHEN** user selects a done or dropped project
- **THEN** the delete keybinding SHALL be disabled

### Requirement: Park project with "s" key
The system SHALL allow parking an open project by pressing "s" (someday), transitioning it immediately without confirmation.

#### Scenario: Park open project
- **WHEN** user selects an open project
- **AND** presses "s"
- **THEN** ParkProject SHALL be called
- **AND** the list SHALL reload

#### Scenario: Park disabled for non-open
- **WHEN** user selects a someday, done, or dropped project
- **THEN** the "s" keybinding SHALL be disabled

### Requirement: Reorder projects with shift+up/down
The system SHALL allow reordering open and someday projects within their status group using shift+up and shift+down.

#### Scenario: Move project up
- **WHEN** user selects an open project that is not the first open project
- **AND** presses shift+up
- **THEN** MoveProjectUp SHALL be called
- **AND** the list SHALL reload with the cursor on the moved project

#### Scenario: Move project down
- **WHEN** user selects an open project that is not the last open project
- **AND** presses shift+down
- **THEN** MoveProjectDown SHALL be called
- **AND** the list SHALL reload with the cursor on the moved project

#### Scenario: Move disabled at boundary
- **WHEN** user selects the first open project
- **THEN** shift+up SHALL be disabled
- **WHEN** user selects the last open project within its status group
- **THEN** shift+down SHALL be disabled

#### Scenario: Move disabled for done/dropped
- **WHEN** user selects a done or dropped project
- **THEN** shift+up and shift+down SHALL be disabled

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