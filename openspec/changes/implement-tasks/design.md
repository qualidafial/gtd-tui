## Context

The GTD TUI requires Task entity implementation as defined in the foundation specs (`domain-model` and `architecture`). Tasks are the core actionable unit in GTD - they represent single-step work items that can be next actions (do ASAP) or delegated (waiting on someone else).

The foundation specs already define:
- Task fields: Kind, Status, Due, DeferUntil, ProjectID, Assignee
- Service pattern: value semantics, dedicated transition methods
- SQLite conventions: squirrel queries, CHECK constraints, transactions

This design implements those requirements with specific technical decisions.

## Goals / Non-Goals

**Goals:**
- Implement Task domain type with all required fields
- Implement TaskService interface following established patterns
- Implement SQLite store with proper constraints and transactions
- Support deferred task filtering in list operations
- Maintain consistency with existing architecture patterns

**Non-Goals:**
- UI/TUI implementation (separate change)
- Task-Project cascade operations (part of Project implementation)
- Comment attachment to tasks (separate Comment implementation)
- MeetingLink support (separate Meeting implementation)

## Decisions

### Decision 1: Task struct field layout

Task fields in order: ID, Title, Description, Kind, Status, Assignee, Due, DeferUntil, ProjectID, CreatedAt, UpdatedAt.

**Rationale:** Group identity (ID), content (Title, Description), GTD semantics (Kind, Status, Assignee), scheduling (Due, DeferUntil), relationships (ProjectID), and timestamps together. Assignee follows Status since it's only relevant for delegated tasks.

**Alternatives:**
- Alphabetical order: rejected - harder to read and understand field relationships
- Separate struct for GTD fields: rejected - unnecessary complexity for flat entity

### Decision 2: Kind and Status as string types with constants

Use `type TaskKind string` and `type TaskStatus string` with exported constants (e.g., `TaskKindNextAction = "next_action"`).

**Rationale:** String types provide readable database values and JSON serialization. Constants prevent typos. Database CHECK constraints enforce validity at storage layer.

**Alternatives:**
- iota-based enums: rejected - require custom marshaling and less readable in DB
- Separate validation layer: rejected - CHECK constraints handle this at DB level

### Decision 3: Status transitions as dedicated methods

Implement CompleteTask, DropTask, ReopenTask as separate methods rather than allowing status changes in UpdateTask.

**Rationale:** Per architecture spec, this localizes transition logic. Each method can handle its specific validation (e.g., can't reopen a task that isn't done/dropped).

**Alternatives:**
- Status field in UpdateTask params: rejected - violates architecture spec, spreads transition logic

### Decision 4: List filtering via ListOptions struct

```go
type TaskListOptions struct {
    Status     *TaskStatus
    Kind       *TaskKind
    ProjectID  *int64
    IncludeDeferred bool // default false
}
```

**Rationale:** Pointer fields allow distinguishing "not specified" from "filter by nil ProjectID". IncludeDeferred defaults to false so deferred tasks are hidden in normal views.

**Alternatives:**
- Multiple List method variants (ListPending, ListByProject): rejected - combinatorial explosion
- Filter callback function: rejected - can't push filtering to database

### Decision 5: Migration file naming

Use `0002_tasks.sql` following existing `0001_*.sql` pattern.

**Rationale:** Lexicographic ordering ensures proper sequence. Single migration creates tasks table with all constraints.

**Alternatives:**
- Separate migrations for table/indexes/constraints: rejected - unnecessary for initial creation

## Risks / Trade-offs

**[Risk] Assignee field unused for next_action tasks** → Accept nullable string; application logic can validate Assignee is set when Kind is delegated, but DB doesn't enforce this to keep schema simple.

**[Risk] DeferUntil filtering adds complexity to List queries** → Worth it for GTD workflow; keep default behavior (exclude deferred) and explicit opt-in flag.

**[Risk] No cascade on ProjectID foreign key** → Tasks should not be deleted when Project is deleted; Project completion/dropping will handle task status separately.
