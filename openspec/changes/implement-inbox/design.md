## Context

The GTD TUI app has established domain types for Task and Project with SQLite implementations. The inbox is the entry point for the GTD capture workflow - a staging area for unprocessed thoughts before they are clarified into actionable entities. This change introduces the Item entity and InboxService to enable that capture workflow.

The existing codebase demonstrates patterns we will follow:
- Domain types in root package with service interfaces
- SQLite implementations using squirrel query builder
- Migrations as embedded SQL files with CHECK constraints
- Value semantics throughout (no pointers in service returns)

## Goals / Non-Goals

**Goals:**
- Define Item entity with soft-delete lineage via ClarifiedInto pointers
- Implement InboxService with Create, List, Get operations
- Establish migration for items table with proper constraints
- Follow existing architectural patterns from task/project implementations

**Non-Goals:**
- Clarify operations (Discard, ClarifyAsTask, etc.) - these span multiple services and belong in a future change
- TUI integration - this change establishes the service layer only
- The references table — `ClarifiedIntoReferenceID` is added by implement-clarify when that table exists. Someday items are not a separate entity (they are projects with `Status=someday`), so no `ClarifiedIntoSomedayID` is needed

## Decisions

### Decision: ClarifiedInto as multiple nullable FK columns vs polymorphic reference

**Choice**: Multiple nullable FK columns (clarified_into_task_id, clarified_into_project_id, etc.)

**Rationale**: SQLite CHECK constraints can enforce mutual exclusion directly. This matches the existing pattern for multi-target FKs (see ARCHITECTURE.md on meeting_links). A polymorphic approach (type + id columns) would lose FK integrity and complicate queries.

**Alternatives considered**:
- Polymorphic `clarified_into_type` + `clarified_into_id`: Loses foreign key integrity, requires application-level validation
- Separate junction table: Over-engineered for a 0..1 relationship

### Decision: Discarded as boolean vs status enum

**Choice**: Boolean `discarded` column

**Rationale**: Items have exactly two terminal states: clarified into something, or discarded. A status enum (inbox/clarified/discarded) adds complexity without benefit since "clarified" is already expressed by having a ClarifiedInto pointer set. The boolean keeps the schema simple and the constraint easy to express.

**Alternatives considered**:
- Status enum: Would require coordinating enum values with ClarifiedInto presence
- Deleted_at timestamp: Overkill for a simple boolean flag with no audit requirement

### Decision: List ordering by CreatedAt ASC

**Choice**: Oldest items first (FIFO)

**Rationale**: GTD inbox processing is FIFO - you work through items in capture order to maintain cognitive flow. This matches user expectation for inbox processing and is consistent with how email inboxes work.

**Alternatives considered**:
- Most recent first: Would break FIFO processing flow
- Manual ordering via order_key: Premature complexity for inbox items

### Decision: No ClarifiedIntoSomedayID column

**Choice**: Someday items reuse the Project entity (`Status=someday`). Incubate stamps `ClarifiedIntoProjectID`, the same field as ClarifyAsProject. The two operations differ only in the initial project status.

**Rationale**: The Project entity already models someday via `ProjectStatusSomeday` (per the finalized `project-entity` spec), and `ParkProject`/`ReopenProject` already handle transitions in and out of someday. A parallel Someday entity would duplicate that machinery and force users to "promote" someday items into projects manually.

**Alternatives considered**:
- Separate Someday entity with its own table and ClarifiedIntoSomedayID column: Forces a duplicate promotion path and a second list to maintain
- Add a someday column to items: Conflates inbox state with downstream project status

### Decision: Defer references FK column to implement-clarify

**Choice**: `clarified_into_reference_id` is added by the implement-clarify migration alongside the references table, not by this change.

**Rationale**: Adding the column here would require a NULLable column with no REFERENCES clause and CHECK-constraint coordination across two changes. Deferring it to the change that owns the references table keeps the migration self-contained.

**Alternatives considered**:
- Add the column now without REFERENCES: Splits CHECK-constraint maintenance across two migrations
- Create a stub references table: Adds an unused table that may not match final design

## Risks / Trade-offs

**Risk**: implement-clarify must ALTER the items table to add `clarified_into_reference_id` and extend the CHECK constraint.
**Mitigation**: Documented as a follow-up migration step in implement-clarify's tasks; SQLite supports the operation via table-rebuild pattern already used elsewhere.

**Risk**: Large inbox could make List() slow without pagination.
**Mitigation**: For v1, unbounded List is acceptable for personal use. Add pagination or cursor-based iteration if performance becomes an issue.
