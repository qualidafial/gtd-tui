## ADDED Requirements

### Requirement: Reference entity structure
The system SHALL provide a Reference entity representing standalone markdown content kept for retrieval. A Reference SHALL have:
- ID (int64, assigned by database)
- Title (non-empty string)
- Body (markdown string, may be empty)
- CreatedAt (timestamp, assigned on creation)
- UpdatedAt (timestamp, updated on modification)

#### Scenario: Create reference with required fields
- **WHEN** creating a Reference with title "Pasta Recipe"
- **THEN** the system creates a Reference with ID assigned
- **AND** Body defaults to empty string
- **AND** UpdatedAt equals CreatedAt

#### Scenario: Create reference with body
- **WHEN** creating a Reference with title and markdown body
- **THEN** the system creates a Reference with both fields populated

#### Scenario: Reject empty title
- **WHEN** creating a Reference with empty title
- **THEN** the system rejects with validation error

### Requirement: Reference store interface
The system SHALL provide a ReferenceStore interface with CRUD operations. Create and Update SHALL return the entity with server-assigned fields populated.

#### Scenario: Create returns populated entity
- **WHEN** Create is called with a Reference
- **THEN** the returned Reference has ID, CreatedAt, UpdatedAt populated

#### Scenario: Update returns refreshed entity
- **WHEN** Update is called with a modified Reference
- **THEN** the returned Reference has UpdatedAt refreshed

#### Scenario: Get retrieves by ID
- **WHEN** Get is called with a valid ID
- **THEN** the system returns the Reference with that ID

#### Scenario: Get returns error for missing ID
- **WHEN** Get is called with non-existent ID
- **THEN** the system returns an error

#### Scenario: Delete removes entity
- **WHEN** Delete is called with a valid ID
- **THEN** the Reference is removed from storage

#### Scenario: List returns all references
- **WHEN** List is called
- **THEN** the system returns all Reference entities

### Requirement: Reference not linked to projects or tasks
References SHALL NOT have ProjectID or TaskID fields. References are standalone retrieval content, not linked to other entities.

#### Scenario: Reference has no project field
- **WHEN** examining Reference entity
- **THEN** it has no ProjectID field

#### Scenario: Reference has no task field
- **WHEN** examining Reference entity
- **THEN** it has no TaskID field

### Requirement: Reference SQLite storage
The Reference entity SHALL be stored in a SQLite table with CHECK constraint enforcing non-empty title.

#### Scenario: SQLite migration creates table
- **WHEN** migrations run
- **THEN** a reference table exists with id, title, body, created_at, updated_at columns

#### Scenario: CHECK constraint rejects empty title
- **WHEN** inserting a Reference with empty title directly in SQL
- **THEN** database rejects with CHECK constraint violation

### Requirement: Reference value semantics
Reference entities SHALL use value semantics in service interfaces. No pointers to Reference in store interface signatures.

#### Scenario: Store returns value not pointer
- **WHEN** Create is called
- **THEN** it returns Reference, not *Reference
