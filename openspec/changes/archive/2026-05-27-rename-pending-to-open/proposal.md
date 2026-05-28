## Why

The task status `pending` doesn't align with GTD vocabulary. Projects already use `open` for their active status. Renaming task `pending` to `open` unifies the terminology across entities and better reflects the GTD concept of an actionable item that's available to work on.

Additionally, `Task.Assignee` is a `string` with empty-string semantics for "no assignee." This should be a `*string` (nullable) to match the optional-field pattern used by other fields like `Due`, `DeferUntil`, and `ProjectID`. The migration already requires table recreation, so this is a zero-cost addition.

Finally, `Task.Kind` (`TaskKind` enum: `next_action`/`delegated`) is redundant — a non-nil `Assignee` already implies delegated. Dropping `Kind` simplifies the domain, removes a form field, a query key, a CHECK constraint, and a SQLite column. The migration is already recreating the table, so this is the right time.

## What Changes

- **BREAKING**: Rename `TaskStatusPending` constant from `"pending"` to `"open"` throughout the domain, service, query, SQLite, and TUI layers
- Add a SQLite migration to update existing `pending` rows to `open` and change the CHECK constraint
- Update the task query parser to accept `status:open` (drop `status:pending`)
- Update the default task list query from `status:pending ready:now` to `status:open ready:now`
- Update all specs referencing `pending` to use `open`
- **BREAKING**: Change `Task.Assignee` from `string` to `*string`; SQLite column from `NOT NULL DEFAULT ''` to nullable
- Migrate existing empty-string assignee values to NULL
- **BREAKING**: Remove `TaskKind` type, `Task.Kind` field, `TaskFilter.Kind`, and the `kind` query key
- Drop the `kind` column from the tasks table; remove the Kind select field from the task editor

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `task-entity`: TaskStatus enum value changes from `pending` to `open`; Assignee changes from `string` to `*string`; Kind removed
- `task-query`: Query parser accepts `status:open` instead of `status:pending`; `kind` key removed
- `task-list-query-ui`: Default query changes from `status:pending ready:now` to `status:open ready:now`
- `task-sqlite`: Migration to rename stored values, update CHECK constraint, make assignee nullable, and drop kind column
- `task-edit-ui`: Kind select field removed from task editor
- `task-status-ui`: Display label changes from "Pending" to "Open"
- `task-list-presentation`: Status marker changes for open tasks

## Impact

- **Domain**: `task.go` constant rename, Assignee type change to `*string`, Kind type/field/filter removed
- **SQLite**: New migration altering CHECK constraint, updating existing rows, making assignee nullable, dropping kind column
- **Service/TUI**: All assignee comparisons change from `== ""` to `== nil`; Kind references removed
- **Query parser**: `internal/taskquery/taskquery.go` status value mapping
- **TUI**: Default query string in `tui/app.go`, status rendering in tasklist and taskstatus packages
- **Tests**: All test files referencing `TaskStatusPending` or `"pending"` string