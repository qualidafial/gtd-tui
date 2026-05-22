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
- The somedays and references tables - ClarifiedInto FKs for these will be nullable without foreign key constraints until those entities are implemented

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

### Decision: Defer FK constraints for unimplemented entities

**Choice**: ClarifiedIntoSomedayID and ClarifiedIntoReferenceID columns exist but without REFERENCES clause until those tables are created.

**Rationale**: The Item entity design is stable, but Someday and Reference entities are not yet implemented. Including the columns now allows the domain type to be complete while deferring FK integrity until the target tables exist. Migration ordering ensures items table is created before any entity that might reference it.

**Alternatives considered**:
- Add columns later: Would require schema migration coordination and domain type changes
- Create stub tables: Adds unused tables that may not match final design

## Risks / Trade-offs

**Risk**: ClarifiedInto FKs for Someday/Reference lack referential integrity until those entities exist.
**Mitigation**: The CHECK constraint prevents setting these values. When Someday/Reference are implemented, their migrations will add the FK constraints via ALTER TABLE.

**Risk**: Large inbox could make List() slow without pagination.
**Mitigation**: For v1, unbounded List is acceptable for personal use. Add pagination or cursor-based iteration if performance becomes an issue.

**Trade-off**: Items table has columns for entities not yet implemented.
**Accepted**: The alternative (adding columns later) creates more migration complexity and domain type churn. The current approach front-loads the stable design.
