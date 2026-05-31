## ADDED Requirements

### Requirement: Item entity definition
The system SHALL provide an Item entity in the root package representing an unprocessed capture in the inbox. An Item SHALL have:
- `ID` (int64): unique identifier assigned by the database
- `Title` (string): required, non-empty
- `Description` (string): optional, defaults to empty string
- `CreatedAt` (time.Time): timestamp when captured
- `UpdatedAt` (time.Time): timestamp of last modification
- `ClarifiedIntoTaskID` (*int64): nullable FK to Task if clarified as task
- `ClarifiedIntoProjectID` (*int64): nullable FK to Project — covers both ClarifyAsProject (`Status=open`) and Incubate (`Status=someday`)

The `implement-references` change adds a `ClarifiedIntoReferenceID` field; this spec covers the Item state at the end of this change.
- `Discarded` (bool): true if discarded during clarify

All `ClarifiedInto*` pointers SHALL be mutually exclusive with Discarded: at most one can be set.

#### Scenario: Item with no clarification target
- **WHEN** an Item is created via inbox capture
- **THEN** all ClarifiedInto* fields are nil
- **AND** Discarded is false
- **AND** the Item appears in the inbox list

#### Scenario: Item value semantics
- **WHEN** InboxService.Create is called
- **THEN** the returned Item is a value, not a pointer
- **AND** ID, CreatedAt, UpdatedAt are populated by the service

### Requirement: InboxService interface
The system SHALL define an InboxService interface in the root package with the following methods:
- `Create(ctx context.Context, item Item) (Item, error)`: creates a new inbox item
- `List(ctx context.Context) ([]Item, error)`: returns all unclarified items
- `Get(ctx context.Context, id int64) (Item, error)`: retrieves a single item by ID

#### Scenario: Create inbox item
- **WHEN** InboxService.Create is called with title "Call dentist"
- **THEN** a new Item is persisted with the given title
- **AND** the returned Item has ID, CreatedAt, UpdatedAt populated
- **AND** the Item appears in subsequent List calls

#### Scenario: Create inbox item with description
- **WHEN** InboxService.Create is called with title and description
- **THEN** both fields are persisted
- **AND** the returned Item reflects both values

#### Scenario: List returns only unclarified items
- **WHEN** InboxService.List is called
- **THEN** only Items with all ClarifiedInto* fields nil and Discarded=false are returned
- **AND** Items are ordered by CreatedAt ascending (oldest first, FIFO processing)

#### Scenario: List returns empty when inbox is clear
- **WHEN** InboxService.List is called and all items have been clarified or discarded
- **THEN** an empty slice is returned

#### Scenario: Get returns item by ID
- **WHEN** InboxService.Get is called with a valid ID
- **THEN** the matching Item is returned
- **AND** all fields including ClarifiedInto* are populated

#### Scenario: Get returns error for missing item
- **WHEN** InboxService.Get is called with a non-existent ID
- **THEN** an error is returned indicating the item was not found

### Requirement: SQLite items table
The system SHALL create an `items` table via migration with the following schema:
- `id INTEGER PRIMARY KEY`: auto-incrementing identifier
- `title TEXT NOT NULL CHECK (title != '')`: non-empty title
- `description TEXT NOT NULL DEFAULT ''`: optional description
- `created_at DATETIME NOT NULL`: capture timestamp
- `updated_at DATETIME NOT NULL`: last modification timestamp
- `clarified_into_task_id INTEGER REFERENCES tasks(id)`: nullable FK
- `clarified_into_project_id INTEGER REFERENCES projects(id)`: nullable FK (covers both clarify-as-project and incubate; status lives on the project)
- `discarded INTEGER NOT NULL DEFAULT 0`: boolean flag

A CHECK constraint SHALL enforce that at most one of `clarified_into_task_id` / `clarified_into_project_id` is non-null, and `discarded = 0` when either is set. `implement-references` extends this constraint to include `clarified_into_reference_id`.

A CHECK constraint SHALL enforce that at most one of the clarified_into_* columns or discarded is set.

#### Scenario: Insert valid item
- **WHEN** an item is inserted with title "Buy groceries"
- **THEN** the row is created with id assigned
- **AND** created_at and updated_at are populated

#### Scenario: Reject empty title
- **WHEN** an item is inserted with empty title
- **THEN** the database rejects the insert with a CHECK constraint violation

#### Scenario: Mutual exclusion constraint
- **WHEN** an item has both clarified_into_task_id and discarded=1
- **THEN** the database rejects the update with a CHECK constraint violation

### Requirement: SQLite InboxService implementation
The system SHALL implement InboxService in the sqlite package using squirrel for query construction.

#### Scenario: Create uses squirrel insert
- **WHEN** Create is called
- **THEN** the implementation uses sq.Insert to build the query
- **AND** no raw SQL strings are used

#### Scenario: List uses squirrel select with filter
- **WHEN** List is called
- **THEN** the implementation uses sq.Select with WHERE conditions
- **AND** only unclarified, non-discarded items are returned

#### Scenario: Get uses squirrel select
- **WHEN** Get is called
- **THEN** the implementation uses sq.Select with WHERE id = ?
- **AND** the query returns all Item fields including ClarifiedInto pointers

### Requirement: Timestamps use UTC
All timestamps stored in the items table SHALL be in UTC. The SQLite implementation SHALL convert timestamps to UTC before storage and preserve timezone information on retrieval.

#### Scenario: Created item has UTC timestamps
- **WHEN** an Item is created
- **THEN** CreatedAt and UpdatedAt are set to the current time in UTC

#### Scenario: Updated timestamps use UTC
- **WHEN** an Item is modified
- **THEN** UpdatedAt is refreshed to the current time in UTC
