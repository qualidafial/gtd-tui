## MODIFIED Requirements

### Requirement: Item entity definition
The system SHALL provide an Item entity in the root package representing an unprocessed capture in the inbox. An Item SHALL have:
- `ID` (int64): unique identifier assigned by the database
- `Title` (string): required, non-empty
- `Description` (string): optional, defaults to empty string
- `CreatedAt` (time.Time): timestamp when captured
- `UpdatedAt` (time.Time): timestamp of last modification
- `ClarifiedIntoTaskID` (*int64): nullable FK to Task if clarified as task
- `ClarifiedIntoProjectID` (*int64): nullable FK to Project — covers both ClarifyAsProject (`Status=open`) and Incubate (`Status=someday`)
- `ClarifiedIntoReferenceID` (*int64): nullable FK to Reference if filed via FileAsReference
- `Discarded` (bool): true if discarded during clarify

All `ClarifiedInto*` pointers SHALL be mutually exclusive with each other and with Discarded: at most one can be set.

#### Scenario: Item with no clarification target
- **WHEN** an Item is created via inbox capture
- **THEN** all ClarifiedInto* fields are nil
- **AND** Discarded is false
- **AND** the Item appears in the inbox list

#### Scenario: Item value semantics
- **WHEN** InboxService.Create is called
- **THEN** the returned Item is a value, not a pointer
- **AND** ID, CreatedAt, UpdatedAt are populated by the service

### Requirement: SQLite items table
The system SHALL maintain an `items` table whose schema includes:
- `id INTEGER PRIMARY KEY`
- `title TEXT NOT NULL CHECK (title != '')`
- `description TEXT NOT NULL DEFAULT ''`
- `created_at DATETIME NOT NULL`
- `updated_at DATETIME NOT NULL`
- `clarified_into_task_id INTEGER REFERENCES tasks(id)`
- `clarified_into_project_id INTEGER REFERENCES projects(id)`
- `clarified_into_reference_id INTEGER REFERENCES references(id)`
- `discarded INTEGER NOT NULL DEFAULT 0`

A CHECK constraint SHALL enforce that at most one of `clarified_into_task_id` / `clarified_into_project_id` / `clarified_into_reference_id` is non-null, and `discarded = 0` when any is set. This change rebuilds the items table (per SQLite limitations on altering CHECK constraints) to add `clarified_into_reference_id` and extend the constraint.

#### Scenario: Insert valid item
- **WHEN** an item is inserted with title "Buy groceries"
- **THEN** the row is created with id assigned
- **AND** created_at and updated_at are populated

#### Scenario: Reject empty title
- **WHEN** an item is inserted with empty title
- **THEN** the database rejects the insert with a CHECK constraint violation

#### Scenario: Mutual exclusion across all targets
- **WHEN** an item has both `clarified_into_task_id` and `clarified_into_reference_id` set
- **THEN** the database rejects with a CHECK constraint violation

#### Scenario: Existing rows survive rebuild
- **WHEN** the items rebuild migration runs against a database with existing rows
- **THEN** all rows are preserved with the same ids, timestamps, and `clarified_into_*` values
- **AND** the new column defaults to NULL on existing rows
