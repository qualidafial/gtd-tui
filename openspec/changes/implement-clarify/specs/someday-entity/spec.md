## ADDED Requirements

### Requirement: Someday entity structure
The system SHALL provide a Someday entity representing a parked idea not yet fleshed out enough to be a Project. A Someday SHALL have:
- ID (int64, assigned by database)
- Title (non-empty string)
- Description (optional string)
- ReviewedAt (timestamp, defaults to CreatedAt)
- CreatedAt (timestamp, assigned on creation)
- UpdatedAt (timestamp, updated on modification)

#### Scenario: Create someday with required fields
- **WHEN** creating a Someday with title "Learn Rust"
- **THEN** the system creates a Someday with ID assigned
- **AND** ReviewedAt equals CreatedAt
- **AND** UpdatedAt equals CreatedAt

#### Scenario: Create someday with description
- **WHEN** creating a Someday with title and description
- **THEN** the system creates a Someday with both fields populated

#### Scenario: Reject empty title
- **WHEN** creating a Someday with empty title
- **THEN** the system rejects with validation error

### Requirement: Someday store interface
The system SHALL provide a SomedayStore interface with CRUD operations. Create and Update SHALL return the entity with server-assigned fields populated.

#### Scenario: Create returns populated entity
- **WHEN** Create is called with a Someday
- **THEN** the returned Someday has ID, CreatedAt, UpdatedAt populated

#### Scenario: Update returns refreshed entity
- **WHEN** Update is called with a modified Someday
- **THEN** the returned Someday has UpdatedAt refreshed

#### Scenario: Get retrieves by ID
- **WHEN** Get is called with a valid ID
- **THEN** the system returns the Someday with that ID

#### Scenario: Get returns error for missing ID
- **WHEN** Get is called with non-existent ID
- **THEN** the system returns an error

#### Scenario: Delete removes entity
- **WHEN** Delete is called with a valid ID
- **THEN** the Someday is removed from storage

#### Scenario: List returns all somedays
- **WHEN** List is called
- **THEN** the system returns all Someday entities

### Requirement: Someday ReviewedAt for periodic review
The ReviewedAt timestamp SHALL be used to surface stalest items in periodic review. Items with older ReviewedAt values should appear first in review queues.

#### Scenario: ReviewedAt defaults to creation time
- **WHEN** a Someday is created
- **THEN** ReviewedAt is set to the creation timestamp

#### Scenario: ReviewedAt can be updated explicitly
- **WHEN** a user reviews a Someday item
- **THEN** ReviewedAt can be updated to the current timestamp

#### Scenario: List ordered by ReviewedAt
- **WHEN** listing Someday items for review
- **THEN** items with older ReviewedAt appear first

### Requirement: Someday SQLite storage
The Someday entity SHALL be stored in a SQLite table with CHECK constraint enforcing non-empty title.

#### Scenario: SQLite migration creates table
- **WHEN** migrations run
- **THEN** a someday table exists with id, title, description, reviewed_at, created_at, updated_at columns

#### Scenario: CHECK constraint rejects empty title
- **WHEN** inserting a Someday with empty title directly in SQL
- **THEN** database rejects with CHECK constraint violation

### Requirement: Someday value semantics
Someday entities SHALL use value semantics in service interfaces. No pointers to Someday in store interface signatures.

#### Scenario: Store returns value not pointer
- **WHEN** Create is called
- **THEN** it returns Someday, not *Someday
