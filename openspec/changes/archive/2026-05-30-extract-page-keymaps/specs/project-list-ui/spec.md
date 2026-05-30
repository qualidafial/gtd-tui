## MODIFIED Requirements

### Requirement: Enter project view with enter key
The system SHALL allow entering a project's detail view by pressing enter on the selected project, which pushes the project view screen as an overlay. The keymap field carrying this binding SHALL be named `View` (the key remains `enter`).

#### Scenario: Enter project view
- **WHEN** user selects a project
- **AND** presses enter
- **THEN** the project view screen SHALL be pushed as an overlay
- **AND** the project view SHALL display the selected project's details and tasks

#### Scenario: Return from project view
- **WHEN** user presses esc in the project view
- **THEN** the project view SHALL dismiss
- **AND** the project list SHALL reinitialize

#### Scenario: Binding field name reflects action, not key
- **WHEN** code or tests reference the view binding via the projects keymap
- **THEN** the field SHALL be `projects.KeyMap.View`, not `Enter`

### Requirement: Toggle project status with space key
The system SHALL allow toggling project status with space: completing open projects (with confirmation), and reopening someday/done/dropped projects (immediately). The keymap field carrying this binding SHALL be named `ToggleComplete` (the key is `space`; the displayed label flips between `complete` and `reopen` via `SetHelp`).

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

#### Scenario: Binding field name reflects primary action
- **WHEN** code or tests reference the toggle binding via the projects keymap
- **THEN** the field SHALL be `projects.KeyMap.ToggleComplete`, not `Toggle`