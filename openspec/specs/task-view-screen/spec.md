# task-view-screen Specification

## Purpose
TBD - created by archiving change add-task-view. Update Purpose after archive.
## Requirements
### Requirement: Task view shows task header
The task view screen SHALL display a compact header showing non-empty task attributes in order: title, status, project, assignee, due, and description. The project SHALL be shown as the linked project's name resolved via the injected project-name function, and SHALL be omitted when the task is standalone. Attributes that are empty or zero-valued SHALL be hidden. For now the screen is a single bare panel of fields with no tab chrome.

#### Scenario: All attributes present
- **WHEN** the task view is opened for a delegated task with title, status, project, assignee, due, and description set
- **THEN** the header SHALL display all six attributes

#### Scenario: Standalone task with only title and status
- **WHEN** the task view is opened for a standalone task with no assignee, due, or description
- **THEN** the header SHALL display only title and status
- **AND** the project line SHALL be omitted

### Requirement: Edit task from task view
Pressing `e` in the task view SHALL push the task edit overlay for the current task. The edit overlay SHALL be created without a view factory, so saving an existing task returns to the task view rather than pushing a new view.

#### Scenario: Open task editor
- **WHEN** the user presses `e` in the task view
- **THEN** the task edit overlay SHALL be pushed for the current task

### Requirement: Complete or reopen task from task view
Pressing `space` in the task view SHALL push the task status overlay: the complete transition when the task is open, otherwise the reopen transition.

#### Scenario: Complete an open task
- **WHEN** the user presses `space` in the task view for an open task
- **THEN** the task status overlay SHALL be pushed with the complete transition

#### Scenario: Reopen a closed task
- **WHEN** the user presses `space` in the task view for a done or dropped task
- **THEN** the task status overlay SHALL be pushed with the reopen transition

### Requirement: Drop task from task view
Pressing `delete` in the task view SHALL push the task status overlay with the drop transition. The drop binding SHALL be enabled only when the task is open.

#### Scenario: Drop an open task
- **WHEN** the user presses `delete` in the task view for an open task
- **THEN** the task status overlay SHALL be pushed with the drop transition

#### Scenario: Drop disabled for a closed task
- **WHEN** the task view is showing a done or dropped task
- **THEN** the drop binding SHALL be disabled

### Requirement: Assign project from task view
Pressing `p` in the task view SHALL push the project picker overlay for the current task, allowing the task to be assigned to a project or unlinked.

#### Scenario: Open project picker
- **WHEN** the user presses `p` in the task view
- **THEN** the project picker overlay SHALL be pushed for the current task

### Requirement: Convert task to project from task view
Pressing `c` in the task view SHALL push the convert-to-project wizard for the current task. The binding SHALL be enabled only when the task is standalone. On confirmation the wizard SHALL replace the task view with the new project's view (the task no longer exists as a task).

#### Scenario: Convert a standalone task
- **WHEN** the user presses `c` in the task view for a standalone task
- **THEN** the convert-to-project wizard SHALL be pushed

#### Scenario: Convert disabled for a task already in a project
- **WHEN** the task view is showing a task that belongs to a project
- **THEN** the convert-to-project binding SHALL be disabled

### Requirement: Go to linked project from task view
Pressing `g` in the task view SHALL replace the task view with the linked project's view, using `screen.Replace` so the view stack does not deepen. The binding SHALL be enabled only when the task belongs to a project. There is no reciprocal navigation gesture from the project view back to the task.

#### Scenario: Navigate to the linked project
- **WHEN** the user presses `g` in the task view for a task that belongs to a project
- **THEN** the task view SHALL be replaced by that project's view

#### Scenario: Go-to-project disabled for a standalone task
- **WHEN** the task view is showing a standalone task
- **THEN** the go-to-project binding SHALL be disabled

### Requirement: Reload task after returning from an action
The task view's `Init` SHALL re-fetch the task from the service by ID. Because the application re-initializes the revealed screen when an overlay dismisses, returning from any action overlay (edit, status, project picker) SHALL refresh the displayed task without dedicated per-action reload handling.

#### Scenario: Header updates after edit
- **WHEN** the user edits the task title and saves
- **AND** the editor overlay dismisses
- **THEN** the task view header SHALL display the updated title

#### Scenario: Status updates after completion
- **WHEN** the user completes the task from the task view
- **AND** the status overlay dismisses
- **THEN** the task view header SHALL display the updated status

#### Scenario: Project updates after reassignment
- **WHEN** the user assigns the task to a different project via `p`
- **AND** the picker dismisses
- **THEN** the task view header SHALL display the new linked project

### Requirement: Dismiss task view with esc
Pressing esc SHALL dismiss the task view and return to the task list.

#### Scenario: Esc returns to task list
- **WHEN** the user presses esc in the task view
- **THEN** the task view overlay SHALL be dismissed
- **AND** the task list SHALL reinitialize
