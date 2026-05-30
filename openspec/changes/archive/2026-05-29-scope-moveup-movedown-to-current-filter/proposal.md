## Why

Shift+up / shift+down on a project or task currently swaps `order_key` with the immediate neighbor in the **entire** pending/status group, regardless of what the user is actually looking at. When the query bar narrows the list, the visible cursor often does not appear to move at all — the moved item swapped with a hidden neighbor — and several presses are needed before any visible reorder occurs. Scoping the swap to the active filter makes a single press always shift the item one slot within the visible list.

## What Changes

- **BREAKING**: `gtd.TaskService.MoveTaskUp` / `MoveTaskDown` gain a `TaskFilter` parameter. The implementation computes prev/next neighbors among pending tasks matching the filter (status is always forced to open inside the move; the filter narrows further by search, assignee, project_id, date predicates, etc.).
- **BREAKING**: `gtd.ProjectService.MoveProjectUp` / `MoveProjectDown` gain a `ProjectFilter` parameter. The implementation computes prev/next neighbors among same-status projects matching the filter (the moving project's status group always constrains the search; the filter narrows further by search).
- `service.ProjectTaskService.MoveTaskUp` / `MoveTaskDown` inject the wrapper's `projectID` into the filter (mirroring `ListTasks`) so an in-project view always reorders within that project.
- TUI call sites pass the list's current filter: `tui/pages/tasks/tasklist` passes `m.filter`; `tui/pages/projects/projectlist` passes `m.filter`.
- The order_key write path keeps the existing `orderkey.Between` fast path against the filtered neighbors. On exhaustion the renumber fallback acts on the *entire* pending/status group (not just the filtered subset), preserving the relative order of every item; only the moving item may jump several positions in the global ordering to settle into its filtered target slot. Accepted fallout: a move may visibly shift the moved item past several non-filtered items in any list that lacks the filter; no other items are reordered.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `task-service`: add `TaskFilter` parameter to MoveTaskUp/MoveTaskDown; specify filtered-neighbor semantics.
- `project-service`: add `ProjectFilter` parameter to MoveProjectUp/MoveProjectDown; specify filtered-neighbor semantics.
- `project-task-service-wrapper`: MoveTaskUp/MoveTaskDown are no longer pure delegation — they inject `ProjectID` into the filter before delegating.
- `project-list-ui`: shift+up/down passes the list's active project filter.
- `task-status-ui`: shift+up/down on the task list passes the list's active task filter; the "contiguous block of open tasks" wording is scoped to the visible (filtered) tasks.

## Impact

- Code: `task.go`, `project.go` (interface signatures); `sqlite/task.go`, `sqlite/project.go` (shift impl + filtered-order helper); `service/task.go`, `service/project.go`, `service/project_task.go` (delegation); `tui/pages/tasks/tasklist/model.go`, `tui/pages/projects/projectlist.go` (pass filter); tests in `sqlite/`, `service/`, `tui/pages/`.
- APIs: TaskService and ProjectService move-method signatures change — any in-tree caller must be updated. No DB schema or migration change.
- Behavior change visible to the user: shift+up/down always shifts by one within the current view; in the rare key-exhaustion case the moved item may jump several positions in any *unfiltered* view, but the relative order of all other items is preserved.
