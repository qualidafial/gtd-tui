# task-project-restructure-ui Specification

## Purpose
Defines the TUI surfaces that let the user restructure tasks and projects: converting a standalone task into a project, collapsing an empty project back into a task, and linking standalone tasks into a project.

## Requirements

### Requirement: Convert to Project action on the task list
The task list SHALL expose a Convert to Project action for the selected task. The action SHALL be available only when the selected task is standalone (ProjectID is nil). Invoking it SHALL open the convert-to-project wizard.

#### Scenario: Action available for standalone task
- **WHEN** the selected task in the task list is standalone
- **THEN** the Convert to Project action SHALL be available

#### Scenario: Action unavailable for a task in a project
- **WHEN** the selected task already belongs to a project
- **THEN** the Convert to Project action SHALL NOT be available

### Requirement: Convert to Project wizard
The convert-to-project wizard SHALL be a form-first flow that collects the new project's Title, Outcome, and Description and the re-scoped task's Title and Description, then commits via `ConvertTaskToProject` in a single transaction on submit. Because the source task already exists and is durable, the wizard SHALL NOT use an early checkpoint; abandoning the wizard SHALL leave the task unchanged and standalone.

#### Scenario: Fields pre-populated from the source task
- **WHEN** the wizard opens for a standalone task
- **THEN** the project Title and Description SHALL be pre-populated from the task (editable in place)
- **AND** the reframed task Title and Description SHALL be pre-populated from the task (editable in place)
- **AND** the project Outcome SHALL start empty

#### Scenario: Commit on submit
- **WHEN** the user submits the wizard
- **THEN** `ConvertTaskToProject` SHALL be called with the collected project and reframed task
- **AND** on success the wizard SHALL dismiss

#### Scenario: Abandon leaves the task unchanged
- **WHEN** the user cancels the wizard before submitting
- **THEN** no project SHALL be created
- **AND** the task SHALL remain standalone and unchanged

#### Scenario: Commit error displayed
- **WHEN** `ConvertTaskToProject` fails
- **THEN** the error SHALL be displayed in the wizard
- **AND** the wizard SHALL remain open

### Requirement: Convert to Task action on project view and project list
Both the project view and the project list SHALL expose a Convert to Task action for the focused project. The action SHALL be available only when the project is open and has zero tasks of any status. Invoking it SHALL collapse the project into a standalone task via `ConvertProjectToTask`.

#### Scenario: Action available for an empty open project
- **WHEN** the focused project is open and has zero tasks
- **THEN** the Convert to Task action SHALL be available on both project view and project list

#### Scenario: Action unavailable for a non-empty project
- **WHEN** the focused project has one or more tasks of any status
- **THEN** the Convert to Task action SHALL NOT be available

#### Scenario: Action unavailable for a non-open project
- **WHEN** the focused project's status is someday, done, or dropped
- **THEN** the Convert to Task action SHALL NOT be available

#### Scenario: Convert collapses the project
- **WHEN** the user invokes Convert to Task and confirms
- **THEN** `ConvertProjectToTask` SHALL be called for the project
- **AND** on success the project SHALL be removed from the project list
- **AND** the resulting standalone task SHALL be created

### Requirement: Link Task action on project view
The project view SHALL expose a Link Task action that opens the task picker overlay. The action SHALL be available only when at least one standalone open task exists to link; when there are no candidates the action SHALL be disabled. The picker is selection-only — the project view SHALL apply the chosen task by calling `LinkTaskToProject` for the current project when the picker's selection message arrives, and SHALL own any resulting error.

#### Scenario: Action available when candidates exist
- **WHEN** at least one standalone open task exists
- **THEN** the Link Task action SHALL be available

#### Scenario: Action disabled when no candidates
- **WHEN** no standalone open task exists
- **THEN** the Link Task action SHALL be disabled
- **AND** invoking it SHALL NOT open the picker

#### Scenario: Open the task picker
- **WHEN** the user invokes Link Task with candidates available
- **THEN** the task picker overlay SHALL open

#### Scenario: Project view applies the broadcast selection
- **WHEN** the picker emits a selection message and dismisses
- **THEN** the project view SHALL call `LinkTaskToProject` with the selected task's ID and the current project's ID
- **AND** the linked task SHALL appear in the project's task list

#### Scenario: Link error owned by project view
- **WHEN** `LinkTaskToProject` fails after the picker has dismissed
- **THEN** the project view SHALL surface the error
