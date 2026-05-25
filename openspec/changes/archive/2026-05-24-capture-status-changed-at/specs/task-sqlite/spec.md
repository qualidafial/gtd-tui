## MODIFIED Requirements

### Requirement: Tasks table schema
The tasks table SHALL have columns: id (INTEGER PRIMARY KEY), title (TEXT NOT NULL), description (TEXT NOT NULL DEFAULT ''), kind (TEXT NOT NULL DEFAULT 'next_action'), status (TEXT NOT NULL DEFAULT 'pending'), assignee (TEXT NOT NULL DEFAULT ''), due (DATETIME), defer_until (DATETIME), order_key (TEXT), created_at (DATETIME NOT NULL), updated_at (DATETIME NOT NULL), status_changed_at (DATETIME NOT NULL). (The project_id column is added by `implement-projects`.)

#### Scenario: Create tasks table
- **WHEN** migration runs
- **THEN** tasks table is created with all columns including status_changed_at

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

## ADDED Requirements

### Requirement: Status-changed-at migration
A second migration, sqlite/migrations/0002_task_status_changed_at.sql, SHALL add the status_changed_at column to the tasks table as DATETIME NOT NULL with a default of the current UTC time, and SHALL backfill existing rows by setting status_changed_at equal to updated_at.

#### Scenario: Migration adds the column with a default
- **WHEN** the 0002 migration runs on a database with the original tasks table
- **THEN** the tasks table gains a status_changed_at column that is NOT NULL and defaults to the current UTC time

#### Scenario: Migration backfills existing rows
- **WHEN** the 0002 migration runs on a database that already has tasks
- **THEN** each existing task's status_changed_at is set to its updated_at value
</content>
