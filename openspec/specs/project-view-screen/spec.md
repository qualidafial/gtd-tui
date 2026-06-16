# project-view-screen Specification

## Purpose
Defines the behavior of the project view screen, which displays a project's header attributes and an embedded task list scoped to that project.
## Requirements
### Requirement: Project view shows project header
The project view screen SHALL display a compact header showing non-empty project attributes: title, status, outcome, and due. Description SHALL be omitted. Attributes that are empty or zero-valued SHALL be hidden.

#### Scenario: All attributes present
- **WHEN** the project view is opened for a project with title, status, outcome, and due set
- **THEN** the header SHALL display all four attributes

#### Scenario: Only title and status
- **WHEN** the project view is opened for a project with no outcome and no due date
- **THEN** the header SHALL display only title and status

### Requirement: Project view embeds a scoped task list
The project view SHALL embed a `tasklist.Model` below the header, scoped to the project's tasks via a TaskService wrapper. The task list SHALL use an empty default query (showing all tasks regardless of status).

#### Scenario: Task list shows all project tasks
- **WHEN** the project view is opened for a project with pending, done, and dropped tasks
- **THEN** all tasks belonging to the project SHALL be displayed

#### Scenario: Task list does not show tasks from other projects
- **WHEN** the project view is opened
- **THEN** only tasks with a matching ProjectID SHALL appear

### Requirement: Task list receives remaining height
The project view SHALL calculate the header height and pass the remaining screen height to the embedded task list via WindowSizeMsg.

#### Scenario: Header consumes vertical space
- **WHEN** the terminal has height H and the header occupies N lines
- **THEN** the task list SHALL receive a WindowSizeMsg with height H-N

### Requirement: Create task from project view
Pressing `+` or `insert` in the project view SHALL create a new task pre-populated with the project's ID via the scoped TaskService wrapper.

#### Scenario: New task inherits project
- **WHEN** the user presses `+` in the project view
- **THEN** a task edit overlay SHALL be pushed
- **AND** the new task SHALL have ProjectID set to the current project

### Requirement: All task interactions work in project context
The project view SHALL support all existing task interactions: complete, drop, reopen, reorder, edit, query filtering, and project reassignment via `p`. These operate through the embedded tasklist.

#### Scenario: Complete a task in project view
- **WHEN** the user completes a task within the project view
- **THEN** the task is completed and the project view's task list reloads

#### Scenario: Edit a task in project view
- **WHEN** the user edits a task within the project view
- **THEN** the task edit overlay is pushed with the selected task

#### Scenario: Query filters within project
- **WHEN** the user types a query in the project view's task list query bar
- **THEN** results are filtered within the project's tasks only

#### Scenario: Reassign task to different project
- **WHEN** the user presses `p` on a task in the project view
- **THEN** the project picker overlay SHALL be pushed
- **AND** the user MAY assign the task to a different project or unlink it
- **AND** on reload the task SHALL no longer appear in the current project view

### Requirement: Edit project from project view
Pressing `e` in the project view (while the task list is not capturing input) SHALL push the project edit overlay for the current project.

#### Scenario: Open project editor
- **WHEN** the user presses `e` in the project view
- **AND** the task list is not capturing input
- **THEN** the project edit overlay SHALL be pushed with the current project

#### Scenario: Edit key ignored when capturing input
- **WHEN** the user presses `e` while the task list is capturing input (e.g., query bar active)
- **THEN** the keypress SHALL pass through to the task list

### Requirement: Reload project header after edit
When the project edit overlay dismisses, the project view SHALL re-fetch the project from the service and update its header to reflect any changes to title, outcome, or due.

#### Scenario: Header updates after edit
- **WHEN** the user edits the project title and saves
- **AND** the editor overlay dismisses
- **THEN** the project view header SHALL display the updated title

#### Scenario: Header updates after outcome change
- **WHEN** the user edits the project outcome and saves
- **AND** the editor overlay dismisses
- **THEN** the project view header SHALL display the updated outcome

### Requirement: Dismiss project view with esc
Pressing esc SHALL dismiss the project view and return to the project list.

#### Scenario: Esc returns to project list
- **WHEN** the user presses esc in the project view (while not capturing input)
- **THEN** the project view overlay SHALL be dismissed
- **AND** the project list SHALL reinitialize

### Requirement: Open task view from project view
Pressing `enter` on a task in the project view SHALL push that task's view screen, mirroring the top-level Tasks tab. The `e` binding SHALL continue to open the task editor. Because the project is already the parent screen, the task view pushed from a project view SHALL disable the go-to-project (`g`) action.

#### Scenario: Enter opens the selected task's view
- **WHEN** the user presses `enter` on a task in the project view
- **THEN** the task view screen is pushed for the selected task

#### Scenario: Edit still opens the editor
- **WHEN** the user presses `e` on a task in the project view
- **THEN** the task edit overlay is pushed for the selected task

#### Scenario: In-project task view disables go-to-project
- **WHEN** a task view is opened from the project view
- **THEN** the go-to-project (`g`) action SHALL NOT be offered, since the project is the parent screen

