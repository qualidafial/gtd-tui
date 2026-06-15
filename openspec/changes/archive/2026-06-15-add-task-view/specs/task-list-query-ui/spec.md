## ADDED Requirements

### Requirement: Open task view with enter
Pressing enter on a selected task (while the query bar is not capturing input) SHALL push the task view screen for that task. This replaces the previous behavior where enter opened the task editor directly; the editor is now reached with `e`.

#### Scenario: Enter opens the task view
- **WHEN** the user presses enter on a selected task
- **AND** the query bar is not capturing input
- **THEN** the task view screen SHALL be pushed for the selected task

#### Scenario: Enter ignored while query bar is focused
- **WHEN** the user presses enter while the query bar is focused
- **THEN** the keypress SHALL apply the query and SHALL NOT open the task view

### Requirement: Edit task with "e" key
Pressing `e` on a selected task (while the query bar is not capturing input) SHALL push the task edit overlay for that task. The overlay is created without a view factory, so saving returns to the task list.

#### Scenario: Edit selected task
- **WHEN** the user presses `e` on a selected task
- **THEN** the task edit overlay SHALL be pushed for that task

## MODIFIED Requirements

### Requirement: Create task with "+" or "insert" key
The task list SHALL allow creating a new task by pressing "+" or "insert", which pushes the task edit overlay for a new task. The editor SHALL be created with a view factory so that, on successful create, the new task's view screen is shown (the editor dismisses and the task view is pushed).

#### Scenario: Create new task
- **WHEN** user presses "+" or "insert"
- **THEN** a task edit overlay SHALL be pushed for a new task with default values

#### Scenario: New task lands on its view
- **WHEN** the user submits the new-task form successfully
- **THEN** the editor overlay SHALL dismiss
- **AND** the new task's view screen SHALL be pushed