# project-task-relationship Specification

## Purpose
Defines how tasks relate to projects: at most one project per task, status-cascade rules across project transitions, default someday-project exclusion, detach-to-standalone, and the project-to-tasks reverse query.

## Requirements

### Requirement: Task belongs to zero or one projects
A Task SHALL have an optional ProjectID field (nullable int64) that references a Project. A task belongs to zero or one projects. Standalone tasks have nil ProjectID.

#### Scenario: Task with project
- **WHEN** a task is assigned to a project
- **THEN** Task.ProjectID references the project's ID

#### Scenario: Standalone task
- **WHEN** a task has no project assignment
- **THEN** Task.ProjectID is nil

### Requirement: No pending tasks under closed projects
After a project transitions to done or dropped status, no pending tasks SHALL remain under the project. This is an enforced invariant. If it ever breaks, the bug is visible because pending tasks under closed projects are NOT filtered out of views.

#### Scenario: Complete project clears pending tasks
- **WHEN** CompleteProject is called
- **THEN** no pending tasks remain with ProjectID pointing to the project

#### Scenario: Drop project clears pending tasks
- **WHEN** DropProject is called
- **THEN** no pending tasks remain with ProjectID pointing to the project

#### Scenario: Pending task under closed project is visible
- **WHEN** a bug causes a pending task to exist under a done project
- **THEN** the task appears in views (not filtered out)
- **AND** the inconsistency is discoverable

### Requirement: Done and dropped tasks stay attached
When a project is completed or dropped, done and dropped tasks SHALL remain attached to the project as historical record. Only pending tasks are cascaded or detached.

#### Scenario: Done tasks remain on completed project
- **WHEN** CompleteProject is called on a project with done tasks
- **THEN** the done tasks keep their ProjectID unchanged

#### Scenario: Dropped tasks remain on dropped project
- **WHEN** DropProject is called on a project with dropped tasks
- **THEN** the dropped tasks keep their ProjectID unchanged

### Requirement: Someday project tasks excluded by default
Tasks under a someday project SHALL be excluded from default task views. The TaskFilter field IncludeSomedayProjects (default false) controls this: when false, tasks whose project has someday status are filtered out; when true, they are included. Task statuses are never changed by this filtering. Reopening the project restores tasks to default views automatically.

#### Scenario: Tasks under someday project excluded by default
- **WHEN** listing tasks with default filter (IncludeSomedayProjects=false)
- **AND** a task belongs to a project with status someday
- **THEN** the task is excluded from the results

#### Scenario: Tasks under someday project included on request
- **WHEN** listing tasks with IncludeSomedayProjects=true
- **AND** a task belongs to a project with status someday
- **THEN** the task is included in the results

#### Scenario: Reopen restores task visibility
- **WHEN** a project is reopened (status changed from someday to open)
- **THEN** tasks under the project appear in default views again

#### Scenario: Task status unchanged by parking
- **WHEN** ParkProject is called
- **THEN** all task statuses remain unchanged

### Requirement: Cascade completes pending tasks
When CompleteProject or DropProject is called with cascade=true, all pending tasks under the project SHALL be marked with the same terminal status as the project.

#### Scenario: Cascade done to tasks
- **WHEN** CompleteProject is called with cascade=true
- **THEN** all pending tasks under the project are marked done

#### Scenario: Cascade dropped to tasks
- **WHEN** DropProject is called with cascade=true
- **THEN** all pending tasks under the project are marked dropped

### Requirement: Detach makes tasks standalone
When CompleteProject or DropProject is called with cascade=false, all pending tasks under the project SHALL have their ProjectID set to nil, making them standalone tasks.

#### Scenario: Detach sets ProjectID to nil
- **WHEN** CompleteProject is called with cascade=false
- **THEN** all pending tasks have ProjectID set to nil

#### Scenario: Detached tasks are standalone
- **WHEN** tasks are detached from a project
- **THEN** they appear as standalone tasks in views

### Requirement: Project to tasks reverse query
The system SHALL support querying tasks by ProjectID to list all tasks belonging to a project. This is the reverse direction of the Task -> Project relationship.

#### Scenario: List tasks by project
- **WHEN** listing tasks filtered by ProjectID
- **THEN** all tasks with matching ProjectID are returned

#### Scenario: Project shows linked tasks
- **WHEN** viewing a project
- **THEN** the system can retrieve all tasks linked to that project

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