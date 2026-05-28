# task-sqlite Delta Spec

## REMOVED Requirements

### Requirement: Kind CHECK constraint
**Reason**: Kind column dropped — delegated is inferred from non-nil assignee.
**Migration**: Drop the kind column from the tasks table in the table recreation migration.

## MODIFIED Requirements

### Requirement: Tasks table schema
The tasks table SHALL have columns: id (INTEGER PRIMARY KEY), title (TEXT NOT NULL), description (TEXT NOT NULL DEFAULT ''), status (TEXT NOT NULL DEFAULT 'open'), assignee (TEXT), project_id (INTEGER FK), due (DATETIME), defer_until (DATETIME), order_key (TEXT), created_at (DATETIME NOT NULL), updated_at (DATETIME NOT NULL), status_changed_at (DATETIME NOT NULL). The assignee column SHALL be nullable, with NULL meaning no assignee. The kind column SHALL NOT exist.

#### Scenario: Create tasks table
- **WHEN** migration runs
- **THEN** tasks table is created with all columns including status_changed_at

### Requirement: Status CHECK constraint
A CHECK constraint SHALL enforce status IN ('open', 'done', 'dropped').

#### Scenario: Valid status accepted
- **WHEN** inserting a task with status = 'open'
- **THEN** the insert succeeds

#### Scenario: Invalid status rejected
- **WHEN** inserting a task with status = 'invalid'
- **THEN** the insert fails with CHECK constraint violation
