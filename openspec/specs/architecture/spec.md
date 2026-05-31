# architecture Specification

## Purpose
Defines the system's structural and storage conventions: the layered Go package layout (domain → sqlite → service → tui), service-interface design, the SQLite persistence strategy (driver, query builder, migrations, constraints, transactions), and testing conventions. This is the authoritative description of *how the code is organized and how persistence works*, independent of the domain semantics in `domain-model`.
## Requirements
### Requirement: Go project layout
The system SHALL follow Ben Johnson's Go application structure:
- Root package (`gtd`) — domain types and service interfaces only. No I/O, no dependencies.
- `sqlite/` — SQLite implementation of the service interfaces.
- `service/` — cross-service orchestration (e.g. clarify flow spanning Inbox + Task + Project).
- `tui/` — Bubbletea v2 UI; pages under `tui/pages/`.
- `internal/set/` — generic Set type used internally.
- `cmd/` — entry point(s).

#### Scenario: Domain types in root package
- **WHEN** defining a domain type
- **THEN** it SHALL be placed in the root package
- **AND** it SHALL have no I/O or external dependencies

#### Scenario: SQLite implementation in sqlite package
- **WHEN** implementing a service interface
- **THEN** SQLite implementation SHALL be in the sqlite/ package

### Requirement: Service interface design
Service interfaces SHALL be defined in the root package alongside their domain types. Read methods SHALL return values. Write methods SHALL return the updated value with server-assigned fields (ID, CreatedAt, UpdatedAt) populated.

#### Scenario: Create method returns populated value
- **WHEN** CreateTask is called
- **THEN** the returned Task has ID, CreatedAt, UpdatedAt populated

#### Scenario: Update method returns updated value
- **WHEN** UpdateTask is called
- **THEN** the returned Task has UpdatedAt refreshed

### Requirement: Status transitions as dedicated methods
Status transitions SHALL be exposed as per-transition methods (CompleteTask, DropTask, ReopenTask, CompleteProject, etc.) rather than free-form status setters in Update*. This keeps cascade logic localized to one place per transition.

#### Scenario: Complete task method
- **WHEN** completing a task
- **THEN** use CompleteTask method, not UpdateTask with status change

#### Scenario: Complete project with cascade
- **WHEN** completing a project
- **THEN** use CompleteProject(dropOpenTasks, comment) method

### Requirement: Transactional clarify operations
The clarify flow on InboxService (Discard, Incubate, FileAsReference, ClarifyAsTask, ClarifyAsProject) SHALL be transactional: each method creates the destination entity and stamps the Item's ClarifiedInto pointer atomically.

#### Scenario: ClarifyAsTask is atomic
- **WHEN** ClarifyAsTask is called
- **THEN** Task creation and Item.ClarifiedInto update occur in one transaction

### Requirement: SQLite driver selection
The system SHALL use `modernc.org/sqlite` as the SQLite driver — pure Go, no CGO.

#### Scenario: Build without CGO
- **WHEN** building the application
- **THEN** CGO is not required

### Requirement: Squirrel for query construction
The system SHALL use `github.com/Masterminds/squirrel` for all query construction. No raw string SQL outside of migrations.

#### Scenario: Build SELECT query
- **WHEN** constructing a SELECT query
- **THEN** use squirrel builder, not raw SQL strings

### Requirement: Embedded SQL migrations
Migrations SHALL be SQL files embedded via `//go:embed` in `sqlite/migrations/`. They SHALL be named `NNNN_description.sql`, applied in lexicographic order, tracked in a migrations table. Each migration SHALL run in its own transaction.

#### Scenario: Add new migration
- **WHEN** adding a schema change
- **THEN** create a new NNNN_description.sql file in sqlite/migrations/

#### Scenario: Migration applies in transaction
- **WHEN** a migration runs
- **THEN** it executes in its own transaction

### Requirement: Schema constraints via CHECK
CHECK constraints SHALL enforce:
- Non-empty titles where required
- Valid status/kind enum values
- "Exactly one FK set" on dual/multi-nullable-FK tables (comments, meeting_links, items.clarified_into_*)

#### Scenario: Non-empty title constraint
- **WHEN** inserting a Task with empty title
- **THEN** database rejects with CHECK constraint violation

#### Scenario: Exactly one FK constraint
- **WHEN** inserting a Comment with both TaskID and ProjectID set
- **THEN** database rejects with CHECK constraint violation

### Requirement: SQLite pragmas
PRAGMA journal_mode=WAL (concurrent reads) and PRAGMA foreign_keys=ON (enforce ON DELETE CASCADE) SHALL be set on every connection.

#### Scenario: WAL mode enabled
- **WHEN** opening a database connection
- **THEN** journal_mode=WAL is set

#### Scenario: Foreign keys enabled
- **WHEN** opening a database connection
- **THEN** foreign_keys=ON is set

### Requirement: Service-level transactions
Write operations SHALL open a transaction at the service method level and pass it to all internal helpers. Callers see atomic commits or rollbacks. Clarify operations and edit-with-comment operations SHALL be transactional across multiple tables.

#### Scenario: CreateTask is transactional
- **WHEN** CreateTask fails mid-operation
- **THEN** all changes are rolled back

#### Scenario: Edit with comment is atomic
- **WHEN** UpdateTask with comment is called
- **THEN** both Task update and Comment creation are in one transaction

### Requirement: Efficient relationship sync
Relationship sync (e.g., meeting links) SHALL load the current set of link rows from the DB, diff against the desired set, and issue only the inserts and deletes that are necessary. No delete-all + re-insert.

#### Scenario: Update meeting links
- **WHEN** updating a meeting's linked entities
- **THEN** system computes diff and issues minimal changes

### Requirement: Bulk loading to avoid N+1
List endpoints SHALL load relationship rows in a single WHERE id IN (...) query after fetching the primary rows, avoiding N+1 queries.

#### Scenario: List tasks with projects
- **WHEN** listing tasks
- **THEN** project data is loaded in bulk, not per-task

### Requirement: Value semantics and ID conventions
Domain types SHALL use value semantics — no `*Task` or `*Project` in service interfaces. IDs SHALL be `int64` (matching SQLite's `INTEGER PRIMARY KEY`). Nullable timestamps and FKs SHALL use `*T`; timestamps SHALL be stored as UTC. Code SHALL use `new(expr)` (Go 1.26) to take the address of an expression result rather than ad-hoc pointer helpers like `ptrTime`.

#### Scenario: Service returns value, not pointer
- **WHEN** a service method returns a domain type
- **THEN** it returns the value (e.g. `Task`), not a pointer (`*Task`)

#### Scenario: Address of an expression
- **WHEN** code needs a pointer to an expression result
- **THEN** it uses `new(expr)` rather than a helper function

### Requirement: Testing conventions
Tests SHALL use `github.com/stretchr/testify` (assert/require) and prefer table-driven style. Test databases SHALL use `:memory:` via a shared `openTestDB(t)` helper.

#### Scenario: Table-driven test
- **WHEN** writing a test with multiple cases
- **THEN** use table-driven style with testify

#### Scenario: In-memory test database
- **WHEN** a test needs a database
- **THEN** use openTestDB(t) which returns :memory: database

