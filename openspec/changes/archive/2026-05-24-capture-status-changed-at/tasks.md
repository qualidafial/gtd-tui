## 1. Domain

- [x] 1.1 Add `StatusChangedAt time.Time` to `gtd.Task` (after `UpdatedAt`)
- [x] 1.2 Widen the transition methods on the `gtd.TaskService` interface: `CompleteTask`/`DropTask`/`ReopenTask(ctx, id, at time.Time) (Task, error)`

## 2. SQLite

- [x] 2.1 Add migration `sqlite/migrations/0002_task_status_changed_at.sql`. SQLite rejects a non-constant default on `ADD COLUMN` against a non-empty table, so rebuild the table instead: `CREATE TABLE tasks_new` with `status_changed_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))`, `INSERT ... SELECT` copying `updated_at` into `status_changed_at` (backfill), `DROP TABLE tasks`, `RENAME tasks_new TO tasks`, recreate the index
- [x] 2.2 Include `status_changed_at` in the column list, INSERT (CreateTask), and row scans; set it to `created_at` on create
- [x] 2.3 Thread an `at time.Time` param through `transitionTask` and set `status_changed_at = at.UTC()` alongside the existing `updated_at = now()`; update `CompleteTask`/`DropTask`/`ReopenTask` DB wrappers
- [x] 2.4 Update `service/task.go` to pass `at` through to the DB transition methods

## 3. Transition overlay (taskstatus)

- [x] 3.1 Add a `*time.Time` (prefilled with `time.Now()`) to the overlay Model; build a `huh.Form` with a `date.Field` bound to it followed by the existing confirm
- [x] 3.2 Change the `spec.apply` signature to carry the instant; in `applyCmd` resolve an empty/nil timestamp to `time.Now()` before calling the service method
- [x] 3.3 Preserve affirmative-preselect and esc-cancel behavior with the new two-field form

## 4. Relative-time formatter + task editor

- [x] 4.1 Create `internal/reltime` and move `formatWhen` + helpers (`truncateToDay`, `isLocalMidnight`, `formatClock`) there verbatim as exported `reltime.Format`; relocate the formatter unit tests
- [x] 4.2 Update `tui/pages/tasks/tasklist/render.go` to call `reltime.Format`; keep chip tests green
- [x] 4.3 Add a `Status: <Title-cased status> (<reltime.Format(StatusChangedAt, now)>)` line to the editor's read-only header in `taskedit/model.go` (existing-task branch only); add a title-case helper for the status name
- [x] 4.4 Test the editor Status line: pending changed 3 days ago renders `Status: Pending (3d)`; done changed today renders `Status: Done (today)`

## 5. Tests

- [x] 5.1 sqlite: assert `CreateTask` sets `status_changed_at == created_at`; a transition with a backdated instant stores that instant in `status_changed_at` while `updated_at` advances to ~now; a non-status `UpdateTask` leaves `status_changed_at` unchanged
- [x] 5.2 sqlite: migration test — backfill sets `status_changed_at = updated_at` for pre-existing rows
- [x] 5.3 Update existing tests/call sites for the widened transition signatures

## 6. Verification

- [x] 6.1 Run `go test ./...`
- [x] 6.2 Run the TUI: complete a task accepting the default time (Enter-Enter); complete/drop another with a backdated timestamp; reopen one; confirm cancel leaves status unchanged; open an existing task and confirm the `Status: <Status> (<WHEN>)` line
