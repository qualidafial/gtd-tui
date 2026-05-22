## ADDED Requirements

### Requirement: Tasks table schema
The tasks table SHALL have columns: id (INTEGER PRIMARY KEY), title (TEXT NOT NULL), description (TEXT NOT NULL DEFAULT ''), kind (TEXT NOT NULL), status (TEXT NOT NULL DEFAULT 'pending'), assignee (TEXT NOT NULL DEFAULT ''), due (DATETIME), defer_until (DATETIME), project_id (INTEGER REFERENCES projects(id)), created_at (DATETIME NOT NULL), updated_at (DATETIME NOT NULL).

#### Scenario: Create tasks table
- **WHEN** migration runs
- **THEN** tasks table is created with all columns

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

### Requirement: Project foreign key
The project_id column SHALL be a foreign key referencing projects(id). It SHALL allow NULL for standalone tasks. It SHALL NOT cascade on delete (tasks remain when projects are modified).

#### Scenario: Task references valid project
- **WHEN** inserting a task with a valid project_id
- **THEN** the insert succeeds

#### Scenario: Task references invalid project
- **WHEN** inserting a task with a non-existent project_id
- **THEN** the insert fails with foreign key violation

#### Scenario: Standalone task
- **WHEN** inserting a task with NULL project_id
- **THEN** the insert succeeds

### Requirement: TaskStore type
TaskStore SHALL be defined in the sqlite package implementing TaskService interface. It SHALL use squirrel for query construction.

#### Scenario: TaskStore implements TaskService
- **WHEN** creating a TaskStore
- **THEN** it satisfies the TaskService interface

### Requirement: Transactional write operations
All write operations (Create, Update, Complete, Drop, Reopen) SHALL execute within a transaction. If any part fails, the entire operation SHALL roll back.

#### Scenario: Failed create rolls back
- **WHEN** CreateTask fails mid-operation
- **THEN** no partial data is persisted

#### Scenario: Update with comment is atomic
- **WHEN** UpdateTask with comment is called
- **THEN** both task update and comment creation succeed or fail together

### Requirement: Deferred task filtering in queries
ListTasks queries SHALL exclude tasks where defer_until > current time by default. When IncludeDeferred is true, this filter SHALL be removed.

#### Scenario: Default query excludes deferred
- **WHEN** ListTasks is called without IncludeDeferred
- **THEN** SQL WHERE clause includes defer_until IS NULL OR defer_until <= now

#### Scenario: Include deferred removes filter
- **WHEN** ListTasks is called with IncludeDeferred = true
- **THEN** SQL WHERE clause does not filter by defer_until

### Requirement: Migration file
Migration SHALL be in sqlite/migrations/0002_tasks.sql. It SHALL create the tasks table with all columns and constraints in a single migration.

#### Scenario: Migration file location
- **WHEN** looking for tasks migration
- **THEN** it is at sqlite/migrations/0002_tasks.sql

#### Scenario: Migration creates complete table
- **WHEN** migration runs
- **THEN** tasks table has all columns and constraints

### Requirement: Timestamp handling
Timestamps SHALL be stored as UTC. The store SHALL convert time.Time values to UTC on write and parse as UTC on read.

#### Scenario: Store timestamp as UTC
- **WHEN** creating a task with a local timezone
- **THEN** database stores the time in UTC

#### Scenario: Read timestamp as UTC
- **WHEN** reading a task from database
- **THEN** time.Time values are in UTC timezone
