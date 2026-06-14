## Why

The task and project lists let you nudge the selected item one slot at a time with `shift+↑`/`shift+↓`, but there is no way to send an item straight to the top or bottom of its order. Repeatedly tapping shift+arrow to move an item across a long list is tedious. `shift+home`/`shift+end` are the natural keys for "move first/bottom" and complete the reorder family.

## What Changes

- Add `shift+home` ("move first") and `shift+end` ("move last") reorder bindings to the **task list** and the **project list**, alongside the existing `shift+↑`/`shift+↓` move bindings.
- `shift+home` repositions the selected open item ahead of every other item matching the active filter; `shift+end` repositions it after every other matching item. Position is computed against the *filtered* set, consistent with the existing one-slot moves.
- Enable the new bindings under the same per-selection rules as `shift+↑`/`shift+↓`: only for orderable (open) items, and only when the move is not already a no-op (item not already first / last in the filtered set).
- Add service operations `MoveTaskFirst` / `MoveTaskLast` and `MoveProjectFirst` / `MoveProjectLast`, with the same filter-scoping, closed-item rejection, and order-key exhaustion fallback as the existing one-slot moves.

These are additive keybindings and service methods; no existing behavior changes.

## Capabilities

### New Capabilities
<!-- none: this extends existing reorder capabilities -->

### Modified Capabilities
- `task-service`: add `MoveTaskFirst` / `MoveTaskLast` to the task reordering requirement.
- `project-service`: add `MoveProjectFirst` / `MoveProjectLast` to the project reordering requirement.
- `task-status-ui`: extend the task-list reorder bindings with `shift+home`/`shift+end` and their enablement rules.
- `project-list-ui`: extend the project-list reorder bindings (`shift+up`/`shift+down`) with `shift+home`/`shift+end`.

## Impact

- Service: `service/task.go`, `service/project.go`, `service/project_task.go` (wrapper) — new pass-through methods.
- SQLite: `sqlite/task.go`, `sqlite/project.go` — generalize the existing `shiftTask`/equivalent reorder helper to target an absolute position (top/bottom) rather than a relative delta.
- TUI: `tui/pages/tasks/tasklist/keymap.go` + `model.go`, `tui/pages/projects/keymap.go` + `projectlist.go` — new bindings, `moveCmd` extension, and enablement in `updateKeybindings`.
- No schema or migration changes; reuses the existing `order_key` ordering mechanism.
