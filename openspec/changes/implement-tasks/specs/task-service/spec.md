## ADDED Requirements

### Requirement: TaskService interface
TaskService interface SHALL be defined in the root package alongside the Task domain type. It SHALL provide methods for task CRUD operations and status transitions.

#### Scenario: TaskService in root package
- **WHEN** defining TaskService
- **THEN** it is placed in the root package with Task struct

### Requirement: CreateTask method
CreateTask(ctx context.Context, task Task) (Task, error) SHALL create a new task. It SHALL return the created task with server-assigned fields (ID, CreatedAt, UpdatedAt) populated.

#### Scenario: Create task returns populated value
- **WHEN** CreateTask is called with a valid task
- **THEN** returned Task has ID assigned
- **AND** CreatedAt and UpdatedAt are set

#### Scenario: Create task with project
- **WHEN** CreateTask is called with ProjectID set
- **THEN** returned Task has ProjectID preserved

### Requirement: UpdateTask method
UpdateTask(ctx context.Context, task Task, comment string) (Task, error) SHALL update an existing task. It SHALL return the updated task with UpdatedAt refreshed. If comment is non-empty, a Comment entity SHALL be created atomically. UpdateTask SHALL NOT allow Status changes; use transition methods instead.

#### Scenario: Update task refreshes timestamp
- **WHEN** UpdateTask is called
- **THEN** returned Task has UpdatedAt refreshed

#### Scenario: Update task with comment
- **WHEN** UpdateTask is called with a non-empty comment
- **THEN** a Comment entity is created attached to the task
- **AND** both updates occur atomically

#### Scenario: Update task rejects status change
- **WHEN** UpdateTask is called with a different Status
- **THEN** an error is returned

### Requirement: GetTask method
GetTask(ctx context.Context, id int64) (Task, error) SHALL retrieve a task by ID. It SHALL return an error if the task does not exist.

#### Scenario: Get existing task
- **WHEN** GetTask is called with a valid ID
- **THEN** the Task is returned

#### Scenario: Get non-existent task
- **WHEN** GetTask is called with an invalid ID
- **THEN** an error is returned

### Requirement: ListTasks method
ListTasks(ctx context.Context, opts TaskListOptions) ([]Task, error) SHALL retrieve tasks matching the filter criteria. By default, tasks with DeferUntil in the future SHALL be excluded unless IncludeDeferred is true.

#### Scenario: List all pending tasks
- **WHEN** ListTasks is called with Status = TaskStatusPending
- **THEN** only pending tasks are returned
- **AND** deferred tasks are excluded by default

#### Scenario: List tasks including deferred
- **WHEN** ListTasks is called with IncludeDeferred = true
- **THEN** tasks with future DeferUntil are included

#### Scenario: List tasks by project
- **WHEN** ListTasks is called with ProjectID set
- **THEN** only tasks belonging to that project are returned

#### Scenario: List standalone tasks
- **WHEN** ListTasks is called filtering for nil ProjectID
- **THEN** only tasks without a project are returned

### Requirement: TaskListOptions struct
TaskListOptions SHALL have fields: Status (*TaskStatus), Kind (*TaskKind), ProjectID (*int64), IncludeDeferred (bool). Pointer fields distinguish "not filtering" from "filter by this value".

#### Scenario: Filter by status
- **WHEN** Status pointer is non-nil
- **THEN** only tasks with matching status are returned

#### Scenario: No status filter
- **WHEN** Status pointer is nil
- **THEN** tasks of all statuses are returned

### Requirement: CompleteTask method
CompleteTask(ctx context.Context, id int64, comment string) (Task, error) SHALL transition a pending task to done. It SHALL return an error if the task is not pending.

#### Scenario: Complete pending task
- **WHEN** CompleteTask is called on a pending task
- **THEN** Status becomes TaskStatusDone
- **AND** UpdatedAt is refreshed

#### Scenario: Complete non-pending task fails
- **WHEN** CompleteTask is called on a done or dropped task
- **THEN** an error is returned

#### Scenario: Complete with comment
- **WHEN** CompleteTask is called with a non-empty comment
- **THEN** a Comment entity is created recording the transition reason

### Requirement: DropTask method
DropTask(ctx context.Context, id int64, comment string) (Task, error) SHALL transition a pending task to dropped. It SHALL return an error if the task is not pending.

#### Scenario: Drop pending task
- **WHEN** DropTask is called on a pending task
- **THEN** Status becomes TaskStatusDropped
- **AND** UpdatedAt is refreshed

#### Scenario: Drop non-pending task fails
- **WHEN** DropTask is called on a done or dropped task
- **THEN** an error is returned

### Requirement: ReopenTask method
ReopenTask(ctx context.Context, id int64, comment string) (Task, error) SHALL transition a done or dropped task back to pending. It SHALL return an error if the task is already pending.

#### Scenario: Reopen done task
- **WHEN** ReopenTask is called on a done task
- **THEN** Status becomes TaskStatusPending
- **AND** UpdatedAt is refreshed

#### Scenario: Reopen dropped task
- **WHEN** ReopenTask is called on a dropped task
- **THEN** Status becomes TaskStatusPending
- **AND** UpdatedAt is refreshed

#### Scenario: Reopen pending task fails
- **WHEN** ReopenTask is called on a pending task
- **THEN** an error is returned
