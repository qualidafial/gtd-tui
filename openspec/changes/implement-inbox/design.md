## Context

The GTD TUI app has established domain types for Task and Project with SQLite implementations. The inbox is the entry point for the GTD capture workflow — a staging area for unprocessed thoughts that the user then clarifies into Tasks or Projects. This change ships both halves of that loop: the Item entity + capture API, and the InboxService that orchestrates four clarify operations (Discard, Incubate, ClarifyAsTask, ClarifyAsProject). Reference handling (the Reference entity and FileAsReference operation) is a follow-up scope owned by `implement-references`.

The existing codebase demonstrates patterns we will follow:
- Domain types in root package with service interfaces
- SQLite implementations using squirrel query builder
- Migrations as embedded SQL files with CHECK constraints
- Value semantics throughout (no pointers in service returns)
- Service-layer transaction ownership for cross-store operations (per `openspec/specs/architecture/spec.md`)

## Goals / Non-Goals

**Goals:**
- Define the Item entity with capture lineage via ClarifiedInto pointers
- Implement the capture-side InboxService (Create / List / Get) in the SQLite layer
- Implement the service-layer InboxService with four transactional clarify operations (Discard, Incubate, ClarifyAsTask, ClarifyAsProject)
- Establish migration for `items` with proper CHECK constraints
- Surface inbox in the TUI as a new tab

**Non-Goals:**
- Reference entity, ReferenceStore, FileAsReference operation, and references TUI tab — owned by `implement-references`
- MeetingLink rewriting on clarification — owned by `implement-meetings`
- Periodic-review surfacing of stale someday projects (handled in the existing projects-tab query filter)
- Promotion flows beyond what `ReopenProject` already provides

## Decisions

### Decision: ClarifiedInto as multiple nullable FK columns vs polymorphic reference

**Choice**: Multiple nullable FK columns (`clarified_into_task_id`, `clarified_into_project_id`; `clarified_into_reference_id` is added by `implement-references`).

**Rationale**: SQLite CHECK constraints can enforce mutual exclusion directly. Matches the existing pattern for multi-target FKs (see `openspec/specs/architecture/spec.md` "Schema constraints via CHECK"). A polymorphic approach (type + id columns) would lose FK integrity and complicate queries.

**Alternatives considered**:
- Polymorphic `clarified_into_type` + `clarified_into_id`: Loses FK integrity, requires application-level validation
- Separate junction table: Over-engineered for a 0..1 relationship

### Decision: No ClarifiedIntoSomedayID column

**Choice**: Someday items reuse the Project entity (`Status=someday`). Incubate stamps `ClarifiedIntoProjectID`, the same field as ClarifyAsProject. The two operations differ only in the initial project status.

**Rationale**: The Project entity already models someday via `ProjectStatusSomeday` (per the finalized `project-entity` spec), and `ParkProject`/`ReopenProject` already handle transitions in and out of someday. A parallel Someday entity would duplicate that machinery and force users to "promote" someday items into projects manually.

**Alternatives considered**:
- Separate Someday entity with its own table and `ClarifiedIntoSomedayID` column: Forces a duplicate promotion path and a second list to maintain
- Add a someday column to items: Conflates inbox state with downstream project status

### Decision: Discarded as boolean vs status enum

**Choice**: Boolean `discarded` column.

**Rationale**: Items have exactly two terminal states: clarified into something, or discarded. A status enum (inbox / clarified / discarded) adds complexity without benefit since "clarified" is already expressed by having a `ClarifiedInto*` pointer set. The boolean keeps the schema simple and the CHECK constraint easy to express.

**Alternatives considered**:
- Status enum: Would require coordinating enum values with `ClarifiedInto*` presence
- `Deleted_at` timestamp: Overkill for a simple boolean flag with no audit requirement

### Decision: List ordering by CreatedAt ASC

**Choice**: Oldest items first (FIFO).

**Rationale**: GTD inbox processing is FIFO — you work through items in capture order to maintain cognitive flow. Matches user expectation for inbox processing and how email inboxes work.

**Alternatives considered**:
- Most recent first: Breaks FIFO flow
- Manual ordering via order_key: Premature complexity for inbox items

### Decision: Service-layer InboxService for clarify orchestration

**Choice**: Place the clarify-orchestrating InboxService in the `service/` package, not on the SQLite layer. The capture-side `InboxService` (Create/List/Get) lives in the root package as an interface and is implemented by the SQLite layer; the service package's `InboxService` consumes that store plus TaskStore / ProjectStore (`implement-references` extends the dependency set to include ReferenceStore).

**Rationale**: Clarify operations span multiple stores (Item + Task / Project) inside one transaction. The `service/` package exists for cross-store orchestration per `openspec/specs/architecture/spec.md`. Keeping the SQLite store single-table preserves its single responsibility.

**Alternatives considered**:
- Put clarify methods directly on `sqlite.ItemStore`: violates single-responsibility — a store should only manage its own table
- Place clarify operations on each destination store (e.g., `TaskStore.ClarifyFromItem`): scatters orchestration logic and duplicates transaction wiring

### Decision: Transaction coordination via callback

**Choice**: The service-layer InboxService takes a transaction-provider function from the SQLite layer for atomic operations, following the existing pattern.

**Pattern**:
```go
func (s *InboxService) ClarifyAsTask(ctx context.Context, itemID int64, task Task) (Task, error) {
    var result Task
    err := s.withTx(ctx, func(tx *sql.Tx) error {
        // 1. Get item, verify not already clarified
        // 2. Create task
        // 3. Stamp item.ClarifiedIntoTaskID
        return nil
    })
    return result, err
}
```

## Risks / Trade-offs

**Risk**: Large inbox could make List() slow without pagination.
**Mitigation**: For v1, unbounded List is acceptable for personal use. Add pagination or cursor-based iteration if performance becomes an issue.

**Trade-off**: MeetingLink rewriting on clarify is not handled here.
**Accepted**: Meetings are owned by implement-meetings, which already specifies how its MeetingLinks follow Items through clarification (including the Incubate-creates-someday-project case). That coupling is documented in the meetings change, not duplicated here.

## Open Questions

### ClarifyAsProject with a completed first action ("I chose do-it-now for the first task of a new project")

`ClarifyAsTask` already supports the do-it-now path: the spawned Task can start in `Status=done` (see `specs/clarify-operations/spec.md`, "ClarifyAsTask do-it-now creates done task"). The analogous case for projects is unresolved: the user clarifies an Item into a *project* but has already performed the first concrete action, so that first action's Task should be created in `done` status, letting the UI immediately prompt for the actual next action.

Today `ClarifyAsProject` creates only the Project (no spawned Task), so there is no place for a pre-completed first action.

**Open:**
- Is this a parameter on `ClarifyAsProject` (e.g., an optional first-action Task that is created in `done` status in the same transaction), or a follow-on call (`ClarifyAsProject` then a separate `ClarifyAsTask`/`CreateTask` targeting the new project)?
- If a parameter: does `ClarifyAsProject` then return three entities (Item, Project, Task) instead of two, and how does that interact with the "returns both entities" scenario?

Out of scope for this change's specs until resolved; tracked here so the decision lands alongside the rest of the clarify surface.
