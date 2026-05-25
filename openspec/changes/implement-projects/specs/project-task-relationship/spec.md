## ADDED Requirements

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

### Requirement: Park project filters tasks from default views
When a project has someday status, tasks under that project SHALL be filtered from default task views. The task statuses SHALL NOT change; only view filtering is affected. Reopening the project restores tasks to default views automatically.

#### Scenario: Tasks under someday project filtered
- **WHEN** listing tasks with default view filter
- **AND** a task belongs to a project with status someday
- **THEN** the task is excluded from the results

#### Scenario: Reopen restores task visibility
- **WHEN** a project is reopened (status changed from someday to active)
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
