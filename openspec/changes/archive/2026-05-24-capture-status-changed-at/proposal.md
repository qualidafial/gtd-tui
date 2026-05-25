## Why

`updated_at` is record time: it moves on *any* edit (title, due date, reorder), so it cannot answer "when was this task completed/dropped?" The status-transition moment is lost the next time the task is touched. We want a dedicated event-time field that records when the last status transition truly occurred, and that the user can backdate to the real-world time of the transition rather than the moment gtd happened to record it.

## What Changes

- Add a status-neutral `StatusChangedAt time.Time` field to `gtd.Task`, set on creation (= `CreatedAt`, the transition into `pending`) and overwritten on every status transition.
- Widen the transition service methods to carry the event time: `CompleteTask`/`DropTask`/`ReopenTask(ctx, id, at time.Time)`. `updated_at` continues to track record time (always `now()`); `status_changed_at` is the supplied event time.
- Add a `status_changed_at` column via migration `0002`, `NOT NULL DEFAULT now()`, backfilling existing rows to `status_changed_at = updated_at` (best-effort proxy). Implemented as a table rebuild because SQLite forbids a non-constant default on `ADD COLUMN` against a non-empty table.
- The status-transition confirmation overlay gains an editable timestamp field (reusing `tui/components/date.Field`) prefilled with the current time; the user may backdate it. An empty value falls back to `now()` so the common Enter-Enter path is unchanged.
- The task editor's read-only properties header (shown when editing an existing task) gains a `Status: <Status> (<WHEN>)` line — the title-cased status with a relative WHEN of `StatusChangedAt`. The relative-time formatter (currently private to the task list) is extracted to a shared package so both the list and the editor use one vocabulary.
- Document the task editor as a capability: a new `task-edit-ui` spec captures the existing editor behavior (editable fields, new-task defaults, read-only header, save/create/update, save-error retry, back-out) alongside the new Status line.

## Capabilities

### New Capabilities
- `task-edit-ui`: the task editor — editable fields, new-task defaults, the read-only properties header (Task ID / Created / Updated / Status), save create-or-update, save-error retry, and back-out. The Status line shows the relative time since the last status change.

### Modified Capabilities
- `task-entity`: `Task` gains `StatusChangedAt`; timestamp semantics extended.
- `task-service`: `CompleteTask`/`DropTask`/`ReopenTask` take an `at time.Time`; `CreateTask` sets `StatusChangedAt`.
- `task-sqlite`: tasks table gains `status_changed_at`; transition writes set it from the supplied instant; new migration `0002` adds and backfills the column.
- `task-status-ui`: the confirmation overlay carries an editable, now-defaulted transition timestamp.

## Impact

- `task.go` — add `StatusChangedAt`; widen the three transition method signatures on `gtd.TaskService`.
- `service/task.go` — thread `at` through the three transition methods.
- `sqlite/task.go` — `transitionTask` sets `status_changed_at` from `at`; `CreateTask`/scans include the column.
- `sqlite/migrations/0002_task_status_changed_at.sql` (new) — add column + backfill.
- `tui/pages/tasks/taskstatus/` — overlay becomes a `huh.Form` with a `date.Field` + confirm; `apply` carries the chosen instant.
- `internal/reltime/` (new) — the relative-time WHEN formatter extracted from `tui/pages/tasks/tasklist/render.go` (no behavior change); `tasklist` and `taskedit` both depend on it.
- `tui/pages/tasks/taskedit/model.go` — add the `Status` line to the read-only header using the shared formatter.
- Out of scope: rendering a "done/dropped <when>" chip from `StatusChangedAt` in the list; a full transition/activity log (deferred to `implement-timelines`); a `ctrl+enter` submit-from-any-field hotkey.
</content>
