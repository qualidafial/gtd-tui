## ADDED Requirements

### Requirement: Free-text LIKE filtering
ListTasks SHALL apply each TaskFilter.Search term as a case-insensitive match against title, description, and assignee. A task matches a term when the term is a substring of any of those three columns. Multiple terms SHALL be ANDed.

#### Scenario: Single term matches across columns
- **WHEN** ListTasks is called with Search = ["bob"]
- **THEN** the WHERE clause matches tasks where lower(title), lower(description), or lower(assignee) contains "bob"

#### Scenario: Multiple terms are ANDed
- **WHEN** ListTasks is called with Search = ["report", "bob"]
- **THEN** only tasks matching both terms are returned

### Requirement: Assignee filtering
ListTasks SHALL apply TaskFilter.Assignee as a case-insensitive substring match against the assignee column.

#### Scenario: Assignee narrows results
- **WHEN** ListTasks is called with Assignee = "bob"
- **THEN** only tasks whose assignee contains "bob" (case-insensitive) are returned

### Requirement: Date-predicate filtering
ListTasks SHALL translate Due, Ready, and Defer DatePredicates into SQL constraints. Time-based predicates resolve to a UTC timestamp (end-of-local-day, except `now` which is the current instant). The mapping SHALL be:

- Due (OnOrBefore): `due IS NOT NULL AND due <= t`
- Ready (AvailableAsOf): `defer_until IS NULL OR defer_until <= t`
- Defer (After): `defer_until > t`
- IsNull: `column IS NULL`; IsNotNull: `column IS NOT NULL`

#### Scenario: Due is cumulative
- **WHEN** ListTasks is called with Due resolved to end-of-day today
- **THEN** the WHERE clause selects rows where due IS NOT NULL AND due <= that UTC timestamp (overdue + due-today)

#### Scenario: Ready includes null and opened gates
- **WHEN** ListTasks is called with Ready resolved to now
- **THEN** the WHERE clause selects rows where defer_until IS NULL OR defer_until <= now

#### Scenario: Defer is strict lower bound
- **WHEN** ListTasks is called with Defer resolved to end-of-day +2
- **THEN** the WHERE clause selects rows where defer_until > that UTC timestamp

#### Scenario: Null and not-null variants
- **WHEN** ListTasks is called with Defer = IsNull (or IsNotNull)
- **THEN** the WHERE clause selects rows where defer_until IS NULL (or IS NOT NULL)

## REMOVED Requirements

### Requirement: Deferred task filtering in queries
**Reason**: Replaced by explicit Ready/Defer date predicates. ListTasks no longer performs implicit deferral filtering; the "available now" behavior is expressed by the caller as `Ready = AvailableAsOf(now)` (the default view's query is `status:pending ready:now`).
**Migration**: Callers that relied on the old default (hide future-deferred) MUST add a Ready predicate (`ready:now` in query syntax). The removed `IncludeDeferred` flag has no replacement — omit any deferral predicate to show all tasks.
