## Why

The `implement-tasks` change moved status out of the task edit form — correct GTD modeling, since status is a transition rather than a free-form field — but only wired `DropTask` into the TUI. `CompleteTask` and `ReopenTask` exist and are tested in the service layer yet are unreachable from the UI, so a user currently cannot mark a task done or reopen a closed task. This was missed scope from `implement-tasks`.

## What Changes

- Add status-transition keybindings to the task list, dependent on the selected task's status:
  - **pending**: `space` → Complete (→ done), `delete` → Drop (→ dropped)
  - **done**: `space` → Reopen (→ pending), `delete` → Drop (→ dropped)
  - **dropped**: `space` → Reopen (→ pending), `delete` → no-op (already dropped)
- `space` acts as a toggle: pending completes; done and dropped both reopen to pending.
- Every status change routes through a confirmation overlay before the transition fires. The overlay preselects the affirmative button so a single Enter confirms.
- The move bindings (shift+up / shift+down) are limited to pending tasks: hidden from the help bar and inert when a done or dropped task is selected, since closed tasks sort to the bottom and are not reorderable.
- Generalize the existing `taskdelete` confirm overlay into a single parameterized transition-confirm overlay covering Complete / Drop / Reopen. It selects the title, description, affirmative label, and service method from the target transition. This **replaces** `taskdelete` rather than cloning it, and is the seam where a future comment-on-transition field will be added. No comment field is built in this change.
- The `space` help-bar label is computed from the selected task's status: `space: complete` when pending, `space: reopen` when done or dropped.
- After a transition, a task that no longer matches the active filter simply disappears from the list; no strike-through or lingering.

## Capabilities

### New Capabilities
- `task-status-ui`: Status-transition keybindings on the task list and the shared confirmation overlay that performs Complete / Drop / Reopen.

### Modified Capabilities
<!-- None: task-list-query-ui is scoped to filtering; no existing spec requirements change. -->

## Impact

- `tui/pages/tasks/tasklist/keymap.go`: add the `space` toggle binding; status-dependent help labels.
- `tui/pages/tasks/tasklist/model.go`: route `space` and `delete` to the confirm overlay by transition; compute the contextual help label.
- `tui/pages/tasks/taskdelete/`: generalized into a parameterized transition-confirm overlay (replaces the drop-only overlay).
- No domain or service changes: `CompleteTask`, `DropTask`, and `ReopenTask` already exist and are tested.
