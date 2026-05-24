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

#### Scenario: Create delegated task with assignee
- **WHEN** CreateTask is called with Kind = TaskKindDelegated and an Assignee
- **THEN** returned Task has Kind and Assignee preserved

### Requirement: UpdateTask method
UpdateTask(ctx context.Context, task Task) (Task, error) SHALL update an existing task. It SHALL return the updated task with UpdatedAt refreshed. UpdateTask SHALL NOT allow Status changes; use transition methods instead. (A `comment` parameter for recording why an edit was made is added by `implement-comments`, not this change.)

#### Scenario: Update task refreshes timestamp
- **WHEN** UpdateTask is called
- **THEN** returned Task has UpdatedAt refreshed

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
ListTasks(ctx context.Context, filter TaskFilter) ([]Task, error) SHALL retrieve tasks matching the filter criteria. By default, tasks with DeferUntil in the future SHALL be excluded unless IncludeDeferred is true.

#### Scenario: List all pending tasks
- **WHEN** ListTasks is called with Status = TaskStatusPending
- **THEN** only pending tasks are returned
- **AND** deferred tasks are excluded by default

#### Scenario: List tasks including deferred
- **WHEN** ListTasks is called with IncludeDeferred = true
- **THEN** tasks with future DeferUntil are included

#### Scenario: List tasks by kind
- **WHEN** ListTasks is called with Kind set
- **THEN** only tasks of that kind are returned

### Requirement: TaskFilter struct
TaskFilter SHALL have fields: Status (*TaskStatus), Kind (*TaskKind), IncludeDeferred (bool), TaskIDs ([]int64). Pointer fields distinguish "not filtering" from "filter by this value". Chained builder methods (WithStatus, WithKind, WithTaskIDs) return a copy with the field set. (A ProjectID filter field is added by `implement-projects`.)

#### Scenario: Filter by status
- **WHEN** Status pointer is non-nil
- **THEN** only tasks with matching status are returned

#### Scenario: No status filter
- **WHEN** Status pointer is nil
- **THEN** tasks of all statuses are returned

### Requirement: CompleteTask method
CompleteTask(ctx context.Context, id int64) (Task, error) SHALL transition a pending task to done. It SHALL return an error if the task is not pending. (A `comment` parameter recording the transition reason is added by `implement-comments`.)

#### Scenario: Complete pending task
- **WHEN** CompleteTask is called on a pending task
- **THEN** Status becomes TaskStatusDone
- **AND** UpdatedAt is refreshed

#### Scenario: Complete non-pending task fails
- **WHEN** CompleteTask is called on a done or dropped task
- **THEN** an error is returned

### Requirement: DropTask method
DropTask(ctx context.Context, id int64) (Task, error) SHALL transition a pending task to dropped. It SHALL return an error if the task is not pending. (A `comment` parameter is added by `implement-comments`.)

#### Scenario: Drop pending task
- **WHEN** DropTask is called on a pending task
- **THEN** Status becomes TaskStatusDropped
- **AND** UpdatedAt is refreshed

#### Scenario: Drop non-pending task fails
- **WHEN** DropTask is called on a done or dropped task
- **THEN** an error is returned

### Requirement: ReopenTask method
ReopenTask(ctx context.Context, id int64) (Task, error) SHALL transition a done or dropped task back to pending. It SHALL return an error if the task is already pending. (A `comment` parameter is added by `implement-comments`.)

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
