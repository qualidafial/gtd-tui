## MODIFIED Requirements

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
</content>
