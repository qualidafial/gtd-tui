## Context

Both task and project lists use a query bar (`tui/components/querybar`) that parses into a typed filter (`TaskFilter` / `ProjectFilter`) and reloads the list. Reordering, however, is filter-blind: `sqlite/task.go::shiftTask` loads *all* pending tasks via `pendingOrder`, and `sqlite/project.go::shiftProject` loads *all* same-status projects via `projectOrderByStatus`. The position of the moved item is computed against that full set, so the new key inserts between two tasks/projects that may be hidden from the user, and the visible list does not appear to change.

Fractional `order_key`s (package `orderkey`) are already in place, with `orderkey.Between(prev, next)` returning a key strictly between two existing keys and a renumber fallback when keys are exhausted.

## Goals / Non-Goals

**Goals:**
- A single shift+up or shift+down always produces a visible one-slot move in whatever list the user is looking at.
- Same wire-up applies to both projects (in the projects tab) and tasks (in both the global tasks tab and the in-project task list).
- Keep the existing `orderkey.Between` fast path and per-group renumber fallback; do not change the storage schema.

**Non-Goals:**
- Re-deriving a globally consistent ordering after a filtered move (we explicitly accept that non-filtered items may interleave).
- Avoiding `order_key` collisions across the filter boundary (also accepted).
- Persisting "view identity" so a filtered ordering can be remembered/restored.
- Batched multi-move (shift+up still moves one slot at a time).

## Decisions

### Compute neighbors from the filtered, pending/same-status set

`shiftTask(ctx, id, delta, filter)` will:
1. Load the moving task; reject if not open.
2. Load `filteredOrder := pending tasks matching filter, excluding id, ordered by (order_key ASC, id ASC)`.
3. Compute `pos` = number of `filteredOrder` entries whose `order_key < movingKey`.
4. `newPos := pos + delta` — no-op if out of range.
5. `prevKey = filteredOrder[newPos-1].key` if newPos>0; `nextKey = filteredOrder[newPos].key` if newPos<len.
6. `orderkey.Between(prevKey, nextKey)` → setOrderKey; on `!ok` renumber the *entire* pending/status group (not just the filtered subset) by walking the full order, slotting the moving id into the position that places it between its filtered prev/next neighbors, and rewriting every order_key with a fresh evenly-spaced sequence.

`shiftProject` is the symmetric change with `ProjectFilter`; same-status remains a hard constraint (the moving project's status group is always the universe).

**Why filtered neighbors (not key-swap):** the user explicitly asked for neighbor-based semantics and accepted that non-filtered items may end up interleaved. Key-swap would guarantee no collisions but each move could perturb the relative order of unrelated items between the swapped pair; neighbor-based is the user's stated preference.

### Force `status = open` for tasks even if filter says otherwise

The TUI only enables move bindings when the selected task is open (see `task-status-ui`), and closed tasks have NULL `order_key`. If a user's filter accidentally narrowed to `status:done`, the move binding would already be disabled at the UI layer, but the SQL helper still forces status=open as a defensive constraint so a misuse of the API can never compute a position against closed tasks.

For projects, the moving project's own status (`open` or `someday`) constrains the group; the user-supplied `ProjectFilter.Status`, if present, must match for the move to even be enabled at the UI layer.

### Extract filter-to-WHERE helpers

`ListTasks` and the move-helpers need the same `TaskFilter` → SQL translation (search, assignee, project_id, date predicates, IncludeSomedayProjects). Factor that into a private helper (`applyTaskListFilter(q sq.SelectBuilder, filter gtd.TaskFilter) sq.SelectBuilder`) shared by both call sites. Same pattern for projects (`applyProjectListFilter`).

### Single `pendingOrder` helper, two call patterns

Rather than maintaining separate `pendingOrder` and `filteredPendingOrder` helpers, extend the existing `pendingOrder(ctx, excludeID)` to take a `gtd.TaskFilter` (with status always overridden to open). `shiftTask` uses it twice:
- Fast path: `pendingOrder(ctx, filter, id)` for filtered neighbors.
- Renumber fallback: `pendingOrder(ctx, gtd.TaskFilter{}, id)` for the full open set.

Both calls share the same helper, so the filter→WHERE wiring lives in exactly one place. Same pattern for projects (`projectOrderByStatus` takes a `ProjectFilter`; the moving project's status is always the universe, and an empty `ProjectFilter` loads the whole same-status group for the renumber fallback).

### `ProjectTaskService.MoveTaskUp/MoveTaskDown` injects ProjectID

`service.ProjectTaskService` already injects its `projectID` into `ListTasks` and `CreateTask`. The Move methods do the same to keep the wrapper's invariant — "this service is scoped to one project" — consistent across all list-like operations.

### TUI call sites

`tasklist.Model.moveCmd` and `projectlist.Model.moveCmd` already capture `m.filter` (used to reload after the move). They will pass that same value into the service call. No new state, no new messages.

## Risks / Trade-offs

- [Items outside the filter may interleave with filtered items after a fast-path move] → Accepted. Documented in proposal and spec. The moving item's new key is placed strictly between its *filtered* prev/next, which may straddle one or more non-filtered items; those items now appear on the opposite side of the moved item in unfiltered views. No other items move.
- [On `orderkey.Between` exhaustion the moving item may visibly jump several positions in unfiltered views] → Accepted. The fallback renumbers the entire pending/status group so the relative order of every non-moving item is preserved; the moving item is slotted between its filtered neighbors, which may be several positions away from its old key in the global ordering. This is preferred over the alternative of renumbering only the filtered subset, which would randomly redistribute filtered items among non-filtered ones.
- [API breaks for any in-tree caller of MoveTaskUp/MoveTaskDown/MoveProjectUp/MoveProjectDown] → Mitigated by being a small, well-contained internal API; every caller is in this repo and will be updated in the same change.
- [Forgetting to pass a filter from a future call site silently falls back to "all open tasks"] → Mitigated by removing the no-filter signature entirely (the parameter is required, callers must pass `gtd.TaskFilter{}` to opt back into the old "global pending" behavior, which makes the intent explicit at the call site).

## Migration Plan

No data migration required. The change is API-only; all in-tree callers are updated in the same commit. There is no out-of-tree consumer.
