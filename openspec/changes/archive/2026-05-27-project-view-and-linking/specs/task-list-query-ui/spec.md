## MODIFIED Requirements

### Requirement: Create task with "+" or "insert" key
The task list SHALL allow creating a new task by pressing "+" or "insert", which pushes the task edit overlay for a new task.

#### Scenario: Create new task
- **WHEN** user presses "+" or "insert"
- **THEN** a task edit overlay SHALL be pushed for a new task with default values

### Requirement: Assign project with "p" key
The task list SHALL allow assigning a project to the selected task by pressing "p", which pushes the project picker overlay via an injected factory function. The task list SHALL NOT import project packages.

#### Scenario: Assign project to task
- **WHEN** user selects a task
- **AND** presses "p"
- **THEN** the project picker overlay SHALL be pushed for the selected task

#### Scenario: Picker factory not provided
- **WHEN** the task list is constructed without a picker factory
- **THEN** the "p" keybinding SHALL be disabled

#### Scenario: Keybinding disabled when no task selected
- **WHEN** the task list is empty
- **THEN** the "p" keybinding SHALL be disabled