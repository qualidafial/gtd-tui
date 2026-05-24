## Context

The GTD TUI requires Task entity implementation as defined in the foundation specs (`domain-model` and `architecture`). Tasks are the core actionable unit in GTD - they represent single-step work items that can be next actions (do ASAP) or delegated (waiting on someone else).

The foundation specs already define:
- Task fields: Kind, Status, Due, DeferUntil, Assignee (ProjectID deferred to `implement-projects`)
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

Task fields in order: ID, Title, Description, Kind, Status, Assignee, Due, DeferUntil, CreatedAt, UpdatedAt.

**Rationale:** Group identity (ID), content (Title, Description), GTD semantics (Kind, Status, Assignee), scheduling (Due, DeferUntil), and timestamps together. Assignee follows Status since it's only relevant for delegated tasks. ProjectID is intentionally NOT part of this change — the task-to-project relationship (field, column, FK) lands in `implement-projects` alongside the projects table it references.

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

### Decision 4: List filtering via TaskFilter with builder methods

```go
type TaskFilter struct {
    Status          *TaskStatus
    Kind            *TaskKind
    IncludeDeferred bool // default false
    TaskIDs         []int64
}

func (f TaskFilter) WithStatus(s TaskStatus) TaskFilter { f.Status = &s; return f }
func (f TaskFilter) WithKind(k TaskKind) TaskFilter     { f.Kind = &k; return f }
func (f TaskFilter) WithTaskIDs(ids ...int64) TaskFilter { f.TaskIDs = ids; return f }
```

**Rationale:** Kept the pre-existing `TaskFilter` name rather than introducing `TaskListOptions`. Pointer fields distinguish "not specified" from "filter by this value". Chained builder methods (`gtd.TaskFilter{}.WithStatus(...)`) read cleanly at call sites and were already the convention in the TUI. `ProjectID` is omitted — it arrives with `implement-projects`. IncludeDeferred defaults to false so deferred tasks are hidden in normal views.

**Alternatives:**
- `TaskListOptions` struct (original plan): rejected - duplicated the existing `TaskFilter` type with a new name
- Multiple List method variants (ListPending, ListByProject): rejected - combinatorial explosion
- Filter callback function: rejected - can't push filtering to database

### Decision 5: Migration strategy — update 0001 in-place

Consolidate the tasks schema into a single migration, `0001_tasks.sql`: corrected status enum (pending/done/dropped), kind + assignee columns, order_key column, and the order_key index. The previously separate `0002_task_order_key.sql` is deleted — its column and index fold into 0001.

**Rationale:** No valuable data exists yet, so the original wrong schema and the separate order_key migration never need to have existed as far as the migration trail is concerned. A single migration keeps the history honest rather than encoding a wrong schema in 0001 and patching it across later files.

**Alternatives:**
- New `0002_tasks.sql` / `0003` revision migration: rejected - would leave 0001 encoding the wrong schema permanently, with no data-preservation benefit to justify it

## Risks / Trade-offs

**[Risk] Assignee field unused for next_action tasks** → Accept nullable string; application logic can validate Assignee is set when Kind is delegated, but DB doesn't enforce this to keep schema simple.

**[Risk] DeferUntil filtering adds complexity to List queries** → Worth it for GTD workflow; keep default behavior (exclude deferred) and explicit opt-in flag.

**[Deferred] ProjectID and its foreign key** → Not part of this change. The field, column, FK (no cascade on delete), and project-status filtering all land in `implement-projects` together with the projects table.

**[Deferred] Comment parameter on UpdateTask and transition methods** → Not part of this change. `UpdateTask`, `CompleteTask`, `DropTask`, and `ReopenTask` ship without a comment parameter; `implement-comments` will modify the task-service capability to add it.
