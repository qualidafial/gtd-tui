## ADDED Requirements

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
