# task-sqlite Specification

## Purpose
TBD - created by syncing change implement-tasks. Update Purpose after archive.

## Requirements

### Requirement: Tasks table schema
The tasks table SHALL have columns: id (INTEGER PRIMARY KEY), title (TEXT NOT NULL), description (TEXT NOT NULL DEFAULT ''), kind (TEXT NOT NULL DEFAULT 'next_action'), status (TEXT NOT NULL DEFAULT 'pending'), assignee (TEXT NOT NULL DEFAULT ''), due (DATETIME), defer_until (DATETIME), order_key (TEXT), created_at (DATETIME NOT NULL), updated_at (DATETIME NOT NULL), status_changed_at (DATETIME NOT NULL). (The project_id column is added by `implement-projects`.)

#### Scenario: Create tasks table
- **WHEN** migration runs
- **THEN** tasks table is created with all columns including status_changed_at

### Requirement: Kind CHECK constraint
A CHECK constraint SHALL enforce kind IN ('next_action', 'delegated').

#### Scenario: Valid kind accepted
- **WHEN** inserting a task with kind = 'next_action'
- **THEN** the insert succeeds

#### Scenario: Invalid kind rejected
- **WHEN** inserting a task with kind = 'invalid'
- **THEN** the insert fails with CHECK constraint violation

### Requirement: Status CHECK constraint
A CHECK constraint SHALL enforce status IN ('pending', 'done', 'dropped').

#### Scenario: Valid status accepted
- **WHEN** inserting a task with status = 'pending'
- **THEN** the insert succeeds

#### Scenario: Invalid status rejected
- **WHEN** inserting a task with status = 'invalid'
- **THEN** the insert fails with CHECK constraint violation

### Requirement: Non-empty title constraint
A CHECK constraint SHALL enforce title != '' (non-empty titles).

#### Scenario: Non-empty title accepted
- **WHEN** inserting a task with a title
- **THEN** the insert succeeds

#### Scenario: Empty title rejected
- **WHEN** inserting a task with empty title
- **THEN** the insert fails with CHECK constraint violation

### Requirement: TaskStore type
TaskStore SHALL be defined in the sqlite package implementing TaskService interface. It SHALL use squirrel for query construction.

#### Scenario: TaskStore implements TaskService
- **WHEN** creating a TaskStore
- **THEN** it satisfies the TaskService interface

### Requirement: Transactional write operations
Write operations that span multiple statements (Update, Complete, Drop, Reopen, and reordering) SHALL execute within a transaction. If any part fails, the entire operation SHALL roll back.

#### Scenario: Failed transition rolls back
- **WHEN** a status transition fails mid-operation
- **THEN** no partial data is persisted

#### Scenario: Status validation and write are atomic
- **WHEN** CompleteTask reads the current status and writes the new one
- **THEN** both occur in a single transaction so a concurrent change cannot slip between them

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

### Requirement: Migration file
The tasks table, all columns and constraints, and the order_key index SHALL all be created by a single migration, sqlite/migrations/0001_tasks.sql.

#### Scenario: Migration file location
- **WHEN** looking for the tasks table migration
- **THEN** it is at sqlite/migrations/0001_tasks.sql

#### Scenario: Migration creates complete table
- **WHEN** the migration runs
- **THEN** tasks table has all columns, constraints, and the order_key index

### Requirement: Timestamp handling
Timestamps SHALL be stored as UTC. The store SHALL convert time.Time values to UTC on write and parse as UTC on read. On a status transition the store SHALL write status_changed_at from the supplied transition instant and refresh updated_at to the current time; on a non-status update it SHALL refresh updated_at and leave status_changed_at unchanged.

#### Scenario: Store timestamp as UTC
- **WHEN** creating a task with a local timezone
- **THEN** database stores the time in UTC

#### Scenario: Read timestamp as UTC
- **WHEN** reading a task from database
- **THEN** time.Time values are in UTC timezone

#### Scenario: Transition writes status_changed_at and updated_at
- **WHEN** a status transition is applied with an instant
- **THEN** status_changed_at is stored as the supplied instant in UTC
- **AND** updated_at is stored as the current time

#### Scenario: Non-status update leaves status_changed_at
- **WHEN** a task field other than status is updated
- **THEN** updated_at is refreshed
- **AND** status_changed_at is unchanged

### Requirement: Status-changed-at migration
A second migration, sqlite/migrations/0002_task_status_changed_at.sql, SHALL add the status_changed_at column to the tasks table as DATETIME NOT NULL with a default of the current UTC time, and SHALL backfill existing rows by setting status_changed_at equal to updated_at.

#### Scenario: Migration adds the column with a default
- **WHEN** the 0002 migration runs on a database with the original tasks table
- **THEN** the tasks table gains a status_changed_at column that is NOT NULL and defaults to the current UTC time

#### Scenario: Migration backfills existing rows
- **WHEN** the 0002 migration runs on a database that already has tasks
- **THEN** each existing task's status_changed_at is set to its updated_at value
