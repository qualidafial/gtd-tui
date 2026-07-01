## Why

Due and defer dates are currently only settable inside the full task editor (`e`), which walks through title, description, assignee, and status before reaching the date fields. Scheduling is a frequent, standalone GTD action — "this is due Friday", "hide this until next week" — and deserves a direct keystroke on the selected task, the same way status changes already have `space`/`delete`.

## What Changes

- Add a `d` binding that opens a date editor for the selected task's **Due** date, and an `f` binding that opens the same editor for its **Defer Until** date.
- Introduce a single small date-editor overlay (one date field + confirm) that prefills the task's current value, accepts natural-language input, clears the date when submitted empty, and commits via `UpdateTask`.
- Wire `d`/`f` on both the task list (including the task list embedded in the project view) and the task view, wherever a single task is focused.
- No change to the project surfaces: Due/DeferUntil are Task-only fields.

## Capabilities

### New Capabilities
- `task-date-ui`: Direct `d` (due) / `f` (defer) keybindings on the task list and task view that open a date-editor overlay to set, change, or clear a task's Due / Defer Until date and commit it via `UpdateTask`.

### Modified Capabilities
<!-- None: date editing is additive. The existing task-list and task-view keymaps
     gain two bindings each, but the new behavior is self-contained in task-date-ui,
     mirroring how task-status-ui owns the status keys + transition overlay. -->

## Impact

- **New package** `tui/pages/tasks/taskdate/`: the date-editor overlay, parameterized by target field (Due | Defer), mirroring the shape of `tui/pages/tasks/taskstatus`.
- **`tui/pages/tasks/tasklist`** (`keymap.go`, `model.go`): two new bindings + two dispatch cases pushing the overlay.
- **`tui/pages/tasks/taskview`** (`keymap.go`, `model.go`): two new bindings + two dispatch cases pushing the overlay.
- **Reuses without change:** `tui/components/form/datefield` (natural-language parsing, empty→nil), `TaskService.UpdateTask`, the `screen.Push` overlay flow.
- **No domain or service changes.** `Task.Due` and `Task.DeferUntil` already exist and are persisted by `UpdateTask`.
- Help bars on the task list and task view each gain two entries (`d due`, `f defer`).
