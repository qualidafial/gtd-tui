# task-entity Delta Spec

## REMOVED Requirements

### Requirement: TaskKind type with constants
**Reason**: Kind is redundant — a non-nil Assignee implies delegated.
**Migration**: Remove all references to TaskKind, TaskKindNextAction, TaskKindDelegated. No replacement needed.

## MODIFIED Requirements

### Requirement: Task struct
The Task struct SHALL be defined in the root package (`gtd`) with the following fields: ID (int64), Title (string), Description (string), Status (TaskStatus), Assignee (*string), ProjectID (*int64), Due (*time.Time), DeferUntil (*time.Time), CreatedAt (time.Time), UpdatedAt (time.Time), StatusChangedAt (time.Time).

#### Scenario: Task struct fields
- **WHEN** inspecting the Task struct
- **THEN** it has all listed fields and no Kind field

### Requirement: TaskStatus type with constants
TaskStatus SHALL be a string type with constants: TaskStatusOpen ("open") for active tasks, TaskStatusDone ("done") for completed tasks, and TaskStatusDropped ("dropped") for abandoned tasks.

#### Scenario: New task is open
- **WHEN** a task is created
- **THEN** Status defaults to TaskStatusOpen

#### Scenario: Complete task
- **WHEN** a task is completed
- **THEN** Status becomes TaskStatusDone

### Requirement: Task nullable fields
Due, DeferUntil, and Assignee fields SHALL be pointer types to represent optional values. Nil indicates not set.

#### Scenario: Task without due date
- **WHEN** creating a task without a due date
- **THEN** Due is nil

#### Scenario: Task without defer date
- **WHEN** creating a task without a defer-until date
- **THEN** DeferUntil is nil

#### Scenario: Task without assignee
- **WHEN** creating a task without an assignee
- **THEN** Assignee is nil

### Requirement: Task timestamps
CreatedAt, UpdatedAt, and StatusChangedAt SHALL be time.Time values stored as UTC. CreatedAt is set on creation. UpdatedAt is set on creation and updated on every modification (record time). StatusChangedAt is set on creation (equal to CreatedAt, representing the transition into open) and updated on every status transition to the instant the transition is recorded as having occurred (event time); it is NOT changed by non-status edits.

#### Scenario: Task creation sets timestamps
- **WHEN** a task is created
- **THEN** CreatedAt and UpdatedAt are both set to current UTC time