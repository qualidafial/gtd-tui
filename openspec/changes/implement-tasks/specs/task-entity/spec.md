## ADDED Requirements

### Requirement: Task struct definition
The Task struct SHALL be defined in the root package (`gtd`) with the following fields: ID (int64), Title (string), Description (string), Kind (TaskKind), Status (TaskStatus), Assignee (string), Due (*time.Time), DeferUntil (*time.Time), ProjectID (*int64), CreatedAt (time.Time), UpdatedAt (time.Time).

#### Scenario: Task struct has all fields
- **WHEN** defining a Task
- **THEN** all fields are present with correct types
- **AND** nullable fields use pointer types

### Requirement: TaskKind type with constants
TaskKind SHALL be a string type with constants: TaskKindNextAction ("next_action") for tasks to do ASAP, and TaskKindDelegated ("delegated") for tasks waiting on someone else.

#### Scenario: Create next action task
- **WHEN** creating a task with Kind = TaskKindNextAction
- **THEN** the task represents work to do ASAP

#### Scenario: Create delegated task
- **WHEN** creating a task with Kind = TaskKindDelegated
- **THEN** the task represents work waiting on someone else
- **AND** the Assignee field identifies who is responsible

### Requirement: TaskStatus type with constants
TaskStatus SHALL be a string type with constants: TaskStatusPending ("pending") for active tasks, TaskStatusDone ("done") for completed tasks, and TaskStatusDropped ("dropped") for abandoned tasks.

#### Scenario: New task is pending
- **WHEN** a task is created
- **THEN** Status defaults to TaskStatusPending

#### Scenario: Complete task
- **WHEN** a task is completed
- **THEN** Status becomes TaskStatusDone

#### Scenario: Drop task
- **WHEN** a task is dropped
- **THEN** Status becomes TaskStatusDropped

### Requirement: Task value semantics
Task SHALL use value semantics throughout service interfaces. No *Task pointers in interface method signatures.

#### Scenario: Service returns Task value
- **WHEN** CreateTask returns
- **THEN** it returns Task, not *Task

### Requirement: Task nullable fields
Due, DeferUntil, and ProjectID fields SHALL be pointer types to represent optional values. Nil indicates not set.

#### Scenario: Task without due date
- **WHEN** creating a task without a due date
- **THEN** Due is nil

#### Scenario: Standalone task
- **WHEN** creating a task not assigned to a project
- **THEN** ProjectID is nil

### Requirement: Task timestamps
CreatedAt and UpdatedAt SHALL be time.Time values stored as UTC. CreatedAt is set on creation. UpdatedAt is set on creation and updated on every modification.

#### Scenario: Task creation sets timestamps
- **WHEN** a task is created
- **THEN** CreatedAt and UpdatedAt are both set to current UTC time

#### Scenario: Task update refreshes UpdatedAt
- **WHEN** a task is updated
- **THEN** UpdatedAt is set to current UTC time
- **AND** CreatedAt remains unchanged
