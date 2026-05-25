## Context

The GTD TUI application needs Projects as a core organizational unit for grouping related tasks under multi-step outcomes. The foundation specs define the Project entity and its relationship to Tasks, but no implementation exists yet.

Current state (verified 2026-05-25):
- Task entity has NO ProjectID field — it was removed in the 05-24 domain reconciliation. This change re-adds it along with the project_id column and migration.
- TaskService exists with CRUD and status transitions. Transitions take an explicit `at time.Time` (e.g. `CompleteTask(ctx, id, at)`) that sets the task's StatusChangedAt.
- Comment entity does NOT exist. It is introduced by `implement-comments`. This change therefore ships comment-free service signatures; comment parameters are added later by that change.
- `project.go`, `project_task.go`, `sqlite/project.go`, `service/project.go` exist only as commented-out scaffolds reflecting a pre-reconciliation design (a `deferred` status, `DeleteProject`, an m:n join table, old method names). They are rewritten, not "uncommented."
- SQLite layer follows established patterns (squirrel, transactions, migrations). Latest migration is `0002_task_status_changed_at.sql`; the projects migration is therefore `0003`.

This change adds the Project entity and ProjectService, following the same patterns established for Tasks.

## Goals / Non-Goals

**Goals:**
- Implement Project domain type in root package
- Implement ProjectService interface with CRUD and status transitions
- Implement SQLite persistence for projects
- Enforce invariant: no pending tasks under closed projects
- Support task filtering by project status (someday filtering)

**Non-Goals:**
- UI/TUI implementation (separate change)
- Project-level timeline/activity views (separate change)
- Meeting-to-project linking via MeetingLink (exists in schema but not this change)
- Comments on projects (owned by `implement-comments`, which lands after this change)

## Decisions

### Decision: Comment parameters deferred to implement-comments

ProjectService methods ship without comment parameters in this change. The Comment entity does not exist yet; `implement-comments` introduces it and re-breaks UpdateProject and the terminal/park transitions to add an optional comment string with atomic Comment creation.

**Rationale:** Projects deliver standalone TUI value without comments. This is a personal project where core-API churn is acceptable, so sequencing projects first (comment-free) and layering comments after is simpler than blocking projects on comments or threading a no-op comment parameter now. See implement-comments `edit-with-comment` spec for the eventual signatures.

### Decision: All transitions thread an `at time.Time` for StatusChangedAt

Every transition — CompleteProject, DropProject, ParkProject, ReopenProject — takes an `at time.Time`. It stamps the project's own `status_changed_at` column, and for cascading transitions (Complete/Drop with cascade=true) the same instant stamps each cascaded task's StatusChangedAt. On CreateProject, `status_changed_at` defaults to `created_at` (the transition into active), exactly as tasks seed StatusChangedAt to CreatedAt.

**Rationale:** Projects record StatusChangedAt for the same reason tasks do (timeline/Reflect views, "in this status since"). Threading an explicit instant — rather than fabricating `time.Now()` internally — lets the caller control the clock and keeps a project and all tasks cascaded in one transition on a single consistent timestamp. This matches the established task-transition convention (`CompleteTask(ctx, id, at)`).

### Decision: Single transition methods over UpdateProject status changes

Status transitions are implemented as dedicated methods (CompleteProject, DropProject, ParkProject, ReopenProject) rather than allowing status changes through UpdateProject.

**Rationale:** Transitions have side effects (cascade/detach tasks, create comments). Centralizing logic in per-transition methods keeps cascade behavior localized and testable. This matches the Task pattern (CompleteTask, DropTask, ReopenTask).

**Alternatives considered:**
- Allow status changes in UpdateProject with hook logic: rejected because it spreads cascade logic and makes testing harder.

### Decision: Cascade flag on terminal transitions

CompleteProject and DropProject take a boolean cascade flag:
- cascade=true: mark all pending tasks with same status
- cascade=false: detach pending tasks (set ProjectID=nil)

**Rationale:** Users need both options. Some projects end with all tasks done; others are abandoned with tasks that should become standalone. Making this explicit per-call avoids implicit behavior that surprises users.

**Alternatives considered:**
- Always cascade: rejected because detaching tasks is a valid use case.
- Always detach: rejected because cascading status is a valid use case.
- Separate DetachTasks method: rejected as unnecessarily verbose when the common case is transition + cascade/detach together.

### Decision: Park/Reopen as status toggle without task cascade

ParkProject sets status to someday; ReopenProject restores a someday/done/dropped project to active. Neither modifies task statuses.

**Rationale:** Parking is reversible and non-destructive. Tasks under parked projects are filtered from default views by query logic, not by status mutation. ReopenProject mirrors ReopenTask, applying equally to parked and terminal (done/dropped) projects.

### Decision: Filter tasks by project status in TaskService

TaskService.ListTasks applies default filtering that excludes tasks whose project has someday status.

**Rationale:** Users don't want parked project tasks cluttering their next actions list. The filtering happens at query time, preserving task state for when the project is reopened.

**Implementation:** JOIN projects and add WHERE clause excluding project.status = 'someday' unless explicitly requested.

### Decision: Projects table schema

```sql
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL CHECK(title != ''),
    outcome TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    due TEXT,  -- ISO8601 timestamp or NULL
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','someday','done','dropped')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    status_changed_at TEXT NOT NULL  -- ISO8601; seeded to created_at on insert
);
```

**Rationale:** Follows existing table patterns (tasks, items). TEXT for timestamps (ISO8601) matches SQLite conventions. CHECK constraints enforce valid status values and non-empty titles.

## Risks / Trade-offs

**[Risk] Orphaned tasks after project deletion** 
Not addressed in this change. Projects are not deleted (only status transitions). Future: if hard delete is added, must handle FK constraint (ON DELETE SET NULL or prevent deletion with pending tasks).

**[Risk] Performance of someday filtering**
Filtering tasks by project status requires a JOIN. For large task/project counts, this could slow down ListTasks.
Mitigation: Add index on projects.status. In practice, personal GTD lists are small enough that this is unlikely to matter.

**[Trade-off] Cascade vs detach is all-or-nothing**
The cascade flag applies to all pending tasks. There's no per-task control.
Mitigation: Users who need partial cascade can detach all, then manually complete specific tasks. This is an acceptable trade-off for simpler API.
