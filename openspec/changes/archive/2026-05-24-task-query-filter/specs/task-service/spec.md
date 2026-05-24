## MODIFIED Requirements

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

## ADDED Requirements

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
