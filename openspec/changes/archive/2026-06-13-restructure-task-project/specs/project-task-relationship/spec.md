## ADDED Requirements

### Requirement: Convert task to project
The system SHALL provide a `ConvertTaskToProject` operation that promotes a standalone task into a new open project, keeping the task as the project's first action. The operation SHALL be transactional. The caller supplies the new project's fields and the re-scoped (reframed) task; the operation owns assigning the task's new ProjectID. Title and description default from the source task when the caller leaves them empty.

#### Scenario: Promote standalone task to project
- **WHEN** `ConvertTaskToProject` is called with a standalone task's ID, project data, and a reframed task
- **THEN** the system creates a Project with `Status=open`
- **AND** the original task's ProjectID is set to the new project's ID
- **AND** the reframed task fields are persisted on the original task
- **AND** the operation returns the created Project and updated Task

#### Scenario: Project fields default from the task
- **WHEN** `ConvertTaskToProject` is called without an explicit project Title or Description
- **THEN** the project Title and Description are copied from the source task

#### Scenario: Atomic conversion
- **WHEN** `ConvertTaskToProject` is called
- **THEN** project creation and the task re-parent/update occur in a single transaction
- **AND** if any step fails, no changes are persisted

#### Scenario: Rejects a task that already belongs to a project
- **WHEN** `ConvertTaskToProject` is called for a task whose ProjectID is non-nil
- **THEN** the system returns an error before any writes occur

### Requirement: Convert empty project to task
The system SHALL provide a `ConvertProjectToTask` operation that collapses an empty open project into a standalone task. The operation SHALL be transactional. It is guarded to projects with `Status=open` that have zero tasks of any status — the only lossless case, since done and dropped tasks are retained as historical record and cannot be collapsed without loss.

#### Scenario: Collapse empty open project to standalone task
- **WHEN** `ConvertProjectToTask` is called for an open project with zero tasks
- **THEN** the system creates a standalone Task with `ProjectID=nil` and `Status=open`
- **AND** the task Title, Description, and Due are copied from the project
- **AND** the project is deleted
- **AND** the operation returns the created Task

#### Scenario: Outcome folded into description
- **WHEN** `ConvertProjectToTask` is called for a project whose Outcome is non-empty
- **THEN** the project Outcome is appended to the resulting task's Description so no content is lost

#### Scenario: Rejects a non-open project
- **WHEN** `ConvertProjectToTask` is called for a project whose status is someday, done, or dropped
- **THEN** the system returns an error before any writes occur

#### Scenario: Rejects a project with any tasks
- **WHEN** `ConvertProjectToTask` is called for a project that has one or more tasks of any status
- **THEN** the system returns an error before any writes occur

#### Scenario: Atomic conversion
- **WHEN** `ConvertProjectToTask` is called
- **THEN** task creation and project deletion occur in a single transaction
- **AND** if any step fails, no changes are persisted

### Requirement: Link standalone task to project
The system SHALL provide a `LinkTaskToProject` operation that re-parents an existing standalone task into a project. The operation SHALL be transactional. Candidates are restricted to standalone tasks (`ProjectID == nil`); the linked task is placed at the bottom of the project's task ordering.

#### Scenario: Link standalone task
- **WHEN** `LinkTaskToProject` is called with a standalone task's ID and a project ID
- **THEN** the task's ProjectID is set to the project's ID
- **AND** the task is ordered at the bottom of the project's tasks
- **AND** the operation returns the updated Task

#### Scenario: Rejects a non-standalone task
- **WHEN** `LinkTaskToProject` is called for a task whose ProjectID is already non-nil
- **THEN** the system returns an error before any writes occur

#### Scenario: Rejects an invalid project
- **WHEN** `LinkTaskToProject` is called with a non-existent project ID
- **THEN** the system returns an error before any writes occur

#### Scenario: Linking into a someday project hides the task from default views
- **WHEN** a standalone task is linked into a project with `Status=someday`
- **THEN** the task's status is unchanged
- **AND** the task is excluded from default task views per the IncludeSomedayProjects rule
- **AND** reopening the project restores the task to default views
