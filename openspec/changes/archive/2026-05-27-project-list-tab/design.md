## Context

The TUI has a single Tasks tab backed by `tasklist.Model`. The tab container already supports multiple tabs with tab/shift+tab switching. The project domain (types, SQLite CRUD, ordering, transitions) is fully implemented but the service layer is commented out, and the `tui/pages/projects/projectlist.go` file is a stub. The task list model provides a proven pattern: `bubbles/list` for rendering, a delegate for row layout, a keymap for actions, status-aware keybinding toggling, and shift+arrows for reordering.

## Goals / Non-Goals

**Goals:**
- Projects tab visible alongside Tasks, navigable via tab/shift+tab
- Project rows show status marker, title, task progress chip, due chip
- Status transitions (complete, drop, park, reopen) via keyboard shortcuts
- Reorder open/someday projects via shift+up/shift+down
- Task progress chip shows pending/total count per project without loading full task objects

**Non-Goals:**
- Project detail view (Change 2)
- Full project edit form (Change 3)
- Project query/filter bar (Change 4)
- Cascade confirmation UI for complete/drop (use cascade=true by default for now; confirm UX deferred)

## Decisions

### Mirror tasklist architecture for projectlist

The project list model follows the same structure as `tasklist.Model`: a `list.Model` from bubbles, a custom delegate for rendering, a keymap struct with per-selection enable/disable, and async commands for loading and reordering. This keeps the codebase consistent and avoids inventing new patterns.

Alternative: a custom rendered list without bubbles. Rejected — bubbles/list provides pagination, scrolling, and status bar for free.

### Task progress via CountTasksByProject query

Add a `CountTasksByProjects(ctx, []int64) (map[int64]ProjectTaskCounts, error)` method on the sqlite layer that returns pending and non-dropped counts per project in a single query. Dropped tasks are excluded from both counts (pending and total). The project list fetches counts for all loaded project IDs in one batch after the project list loads.

Alternative: join counts into the projects query. Rejected — combining the ORDER BY logic with a GROUP BY on the tasks table makes the query fragile. A separate batch query is simpler.

### Status transition overlays reuse taskstatus pattern

Complete and drop transitions push a confirmation overlay (like `taskstatus.Model`) that executes the service call and dismisses. Park, and reopen (from someday/done/dropped) are immediate (no confirmation needed — they're reversible).

Alternative: inline transitions without overlays. Considered but rejected for complete/drop because they have cascade semantics that may surprise users. Park/reopen remain inline because they're trivially undone.

### ProjectService wired as interface

`tui.New` accepts `gtd.ProjectService` alongside `gtd.TaskService`. The project list model receives the `ProjectService` directly. No new domain interface needed — `gtd.ProjectService` already exists.

### Quick-create via "n" key with title-only input

Pressing "n" on the project list pushes a minimal overlay with a single text input for the project title. On submit, it calls `ProjectService.CreateProject` with status=open and the entered title. This is the minimum needed to populate the list for testing. Full editing (outcome, description, due) is deferred to Change 3's edit overlay.

Alternative: inline text input within the list (no overlay). Rejected — the overlay pattern is already established and keeps the list model simple.

### Default list: open projects only

On init, the project list loads `ProjectFilter{Status: &ProjectStatusOpen}`. Someday/done/dropped projects are excluded by default. The query filter (Change 4) will allow switching views later. For now, a simple status filter field on the model could allow toggling to show all.

Simplification: show all projects (no filter) sorted by the existing ORDER BY (open first, then someday, then done/dropped by status_changed_at). This matches user mental model — see everything, act on open ones. Adopt this approach.

## Risks / Trade-offs

- **No cascade confirmation**: complete/drop always cascade (pending tasks get completed/dropped). This matches GTD semantics (closing a project means its tasks are resolved) but could surprise users who want to detach tasks. Mitigation: the overlay shows "N pending tasks will be completed/dropped" text. Full UX deferred to a later change.
- **Batch task counts**: single query for all project IDs. Negligible cost for typical project counts (<50).
- **Minimal creation UX**: quick-create only captures title. Users who want to set outcome/due/description must wait for the edit form (Change 3). Acceptable for incremental rollout — projects are useful with just a title.
