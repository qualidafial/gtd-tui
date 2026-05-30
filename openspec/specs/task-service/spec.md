# task-service Specification

## Purpose
TBD - created by syncing change implement-tasks. Update Purpose after archive.

## Requirements

### Requirement: TaskService interface
TaskService interface SHALL be defined in the root package alongside the Task domain type. It SHALL provide methods for task CRUD operations and status transitions.

#### Scenario: TaskService in root package
- **WHEN** defining TaskService
- **THEN** it is placed in the root package with Task struct

### Requirement: CreateTask method
CreateTask(ctx context.Context, task Task) (Task, error) SHALL create a new task. It SHALL return the created task with server-assigned fields (ID, CreatedAt, UpdatedAt, StatusChangedAt) populated. StatusChangedAt SHALL be set equal to CreatedAt.

#### Scenario: Create task returns populated value
- **WHEN** CreateTask is called with a valid task
- **THEN** returned Task has ID assigned
- **AND** CreatedAt and UpdatedAt are set
- **AND** StatusChangedAt equals CreatedAt

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
ListTasks(ctx context.Context, filter TaskFilter) ([]Task, error) SHALL retrieve tasks matching the filter criteria. ListTasks SHALL perform no implicit deferral filtering: tasks are hidden by deferral only when the caller supplies a Ready or Defer predicate.

#### Scenario: List by status with explicit availability
- **WHEN** ListTasks is called with Status = TaskStatusPending and Ready = AvailableAsOf(now)
- **THEN** only pending tasks that are available now (null or passed defer_until) are returned

#### Scenario: List with no deferral predicate
- **WHEN** ListTasks is called with neither Ready nor Defer set
- **THEN** results are not filtered by defer_until

#### Scenario: List tasks by kind
- **WHEN** ListTasks is called with Kind set
- **THEN** only tasks of that kind are returned

### Requirement: TaskFilter struct
TaskFilter SHALL have fields: Status (*TaskStatus), Kind (*TaskKind), Assignee (*string), Due (*DatePredicate), Ready (*DatePredicate), Defer (*DatePredicate), Search ([]string), TaskIDs ([]int64). It SHALL NOT have an IncludeDeferred field; deferral is expressed with the Ready and Defer predicates. Pointer fields distinguish "not filtering" from "filter by this value" — a nil date predicate means that column is not filtered at all. Chained builder methods (WithStatus, WithKind, WithTaskIDs) return a copy with the field set. (A ProjectID filter field is added by `implement-projects`.)

#### Scenario: Filter by status
- **WHEN** Status pointer is non-nil
- **THEN** only tasks with matching status are returned

#### Scenario: No status filter
- **WHEN** Status pointer is nil
- **THEN** tasks of all statuses are returned

#### Scenario: No deferral filter shows all
- **WHEN** Ready and Defer are both nil
- **THEN** deferral is not filtered at all (deferred and non-deferred tasks are both returned)

#### Scenario: Filter by assignee
- **WHEN** Assignee is non-nil
- **THEN** only tasks whose assignee matches (case-insensitive substring) are returned

#### Scenario: Filter by free-text search
- **WHEN** Search contains one or more terms
- **THEN** only tasks where every term is a case-insensitive substring of the title, description, or assignee are returned

### Requirement: DatePredicate type
A DatePredicate value SHALL carry a kind and (for time-based kinds) a resolved time. Kinds: OnOrBefore, AvailableAsOf, After, IsNull, IsNotNull. ListTasks applies them as: Due uses OnOrBefore (`due ≤ t`, excludes NULL); Ready uses AvailableAsOf (`defer_until IS NULL OR defer_until ≤ t`); Defer uses After (`defer_until > t`, excludes NULL); IsNull/IsNotNull test the column for NULL / NOT NULL.

#### Scenario: Due OnOrBefore predicate
- **WHEN** ListTasks is called with Due = OnOrBefore(today)
- **THEN** tasks due today or earlier (including overdue) are returned

#### Scenario: Ready AvailableAsOf predicate
- **WHEN** ListTasks is called with Ready = AvailableAsOf(now)
- **THEN** tasks with a null defer_until OR defer_until on or before now are returned

#### Scenario: Defer After predicate
- **WHEN** ListTasks is called with Defer = After(+2 days)
- **THEN** only tasks whose defer_until is after that time are returned

#### Scenario: Null and not-null variants
- **WHEN** ListTasks is called with Defer = IsNull (or IsNotNull)
- **THEN** only tasks with a null (or non-null) defer_until are returned

### Requirement: CompleteTask method
CompleteTask(ctx context.Context, id int64, at time.Time) (Task, error) SHALL transition a pending task to done. It SHALL return an error if the task is not pending. The `at` argument is the event time of the transition. (A `comment` parameter recording the transition reason is added by `implement-comments`.)

#### Scenario: Complete pending task
- **WHEN** CompleteTask is called on a pending task with an instant
- **THEN** Status becomes TaskStatusDone
- **AND** StatusChangedAt is set to the supplied instant
- **AND** UpdatedAt is refreshed to current time

#### Scenario: Complete non-pending task fails
- **WHEN** CompleteTask is called on a done or dropped task
- **THEN** an error is returned

### Requirement: DropTask method
DropTask(ctx context.Context, id int64, at time.Time) (Task, error) SHALL transition a pending task to dropped. It SHALL return an error if the task is not pending. The `at` argument is the event time of the transition. (A `comment` parameter is added by `implement-comments`.)

#### Scenario: Drop pending task
- **WHEN** DropTask is called on a pending task with an instant
- **THEN** Status becomes TaskStatusDropped
- **AND** StatusChangedAt is set to the supplied instant
- **AND** UpdatedAt is refreshed to current time

#### Scenario: Drop non-pending task fails
- **WHEN** DropTask is called on a done or dropped task
- **THEN** an error is returned

### Requirement: ReopenTask method
ReopenTask(ctx context.Context, id int64, at time.Time) (Task, error) SHALL transition a done or dropped task back to pending. It SHALL return an error if the task is already pending. The `at` argument is the event time of the transition. (A `comment` parameter is added by `implement-comments`.)

#### Scenario: Reopen done task
- **WHEN** ReopenTask is called on a done task with an instant
- **THEN** Status becomes TaskStatusPending
- **AND** StatusChangedAt is set to the supplied instant
- **AND** UpdatedAt is refreshed to current time

#### Scenario: Reopen dropped task
- **WHEN** ReopenTask is called on a dropped task with an instant
- **THEN** Status becomes TaskStatusPending
- **AND** StatusChangedAt is set to the supplied instant
- **AND** UpdatedAt is refreshed to current time

#### Scenario: Reopen pending task fails
- **WHEN** ReopenTask is called on a pending task
- **THEN** an error is returned

### Requirement: Task reordering
TaskService SHALL provide `MoveTaskUp(ctx context.Context, id int64, filter TaskFilter) error` and `MoveTaskDown(ctx context.Context, id int64, filter TaskFilter) error` to shift an open task one position within the open tasks that match `filter`. The implementation SHALL always constrain the move to status=open regardless of `filter.Status`; the remaining fields of `filter` (ProjectID, Assignee, Search, Due, Ready, Defer, TaskIDs, IncludeSomedayProjects) SHALL narrow the set of candidate neighbors. The move SHALL be rejected for done or dropped tasks. The new position SHALL be computed against the *filtered* set so a single move is one visible slot; items outside the filter MAY interleave with filtered items as a result. On `orderkey.Between` exhaustion, the implementation SHALL renumber the entire set of open tasks (not just the filtered subset), preserving every non-moving task's relative position; only the moving task may visibly jump several positions in unfiltered views.

#### Scenario: Move up within an empty filter
- **WHEN** MoveTaskUp is called on an open task with an empty TaskFilter
- **THEN** the task SHALL move one position earlier among all open tasks

#### Scenario: Move down within a search filter
- **WHEN** MoveTaskDown is called on an open task with a Search filter that matches a subset of open tasks
- **THEN** the task SHALL swap order_keys to land between the next filtered task and the one after it
- **AND** open tasks that do not match the filter SHALL retain their existing order_keys

#### Scenario: Move down with only one matching task is a no-op
- **WHEN** MoveTaskDown is called on an open task that is the only task matching the supplied filter
- **THEN** no order_keys SHALL change

#### Scenario: Reorder rejected for closed task
- **WHEN** MoveTaskUp or MoveTaskDown is called on a done or dropped task
- **THEN** an error SHALL be returned

#### Scenario: Key exhaustion renumbers the entire open set
- **WHEN** MoveTaskUp or MoveTaskDown is called and `orderkey.Between` cannot produce a key strictly between the filtered prev/next neighbors
- **THEN** every open task SHALL be assigned a fresh evenly-spaced order_key in its current order, with the moving task slotted between its filtered neighbors
- **AND** the relative order of every non-moving task SHALL be preserved
