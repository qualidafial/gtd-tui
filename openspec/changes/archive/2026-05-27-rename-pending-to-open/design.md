## Context

Tasks use `pending` as their active status while projects use `open`. This inconsistency creates cognitive friction. The rename is a mechanical find-and-replace across all layers, plus a SQLite migration.

Additionally, `Task.Assignee` uses empty-string semantics for "no assignee" while other optional fields (`Due`, `DeferUntil`, `ProjectID`) use pointer/nil semantics. Since the migration already recreates the table, we normalize assignee to `*string` / nullable at the same time.

`Task.Kind` (`next_action`/`delegated`) is redundant: a non-nil assignee already implies delegated. Dropping it simplifies the domain model, removes a form field, a query key, and a SQLite column. The table recreation makes this free.

## Goals / Non-Goals

**Goals:**
- Rename `TaskStatusPending` / `"pending"` to `TaskStatusOpen` / `"open"` everywhere
- Change `Task.Assignee` from `string` to `*string` with nil meaning "no assignee"
- Remove `TaskKind` type, `Task.Kind` field, `TaskFilter.Kind`, and all references
- Migrate existing SQLite data (status values, empty assignee strings to NULL, drop kind column)
- Keep all existing behavior identical (delegated-ness is inferred from assignee presence)

**Non-Goals:**
- Changing any status transition logic
- Changing project statuses
- Adding any new "delegated" inference logic — just removing the explicit Kind field

## Decisions

### SQLite migration strategy
Use UPDATE + table recreation (same pattern as 0002) to change the CHECK constraint and column nullability. SQLite doesn't support ALTER CHECK or ALTER COLUMN, so we recreate the table. The migration converts `pending` → `open` and empty-string assignee → NULL during the copy.

**Alternative**: Add `open` to the CHECK and leave `pending` as an alias. Rejected — dual values complicate queries and domain logic for no benefit.

### Query parser: clean break
`status:open` replaces `status:pending` with no backward compatibility alias. This is an internal TUI query language with no external consumers.

### Assignee nullability
Change `Task.Assignee` from `string` to `*string`. All `== ""` checks become `== nil`. The `TaskFilter.Assignee` field is already `*string` and needs no change. The SQLite scan uses `sql.NullString` or `*string` to handle NULL. The taskedit form binds to a local `string` and converts to/from `*string` on load/save.

### Dropping Kind
Remove `TaskKind` type, constants, `Task.Kind` field, `TaskFilter.Kind`, `WithKind()`. Remove `kind` from the SQLite column list, CREATE/UPDATE/scan. Remove `kind` from the query parser's recognized keys. Remove the Kind select field from the task editor form. The migration drops the column by omitting it from the new table schema.

## Risks / Trade-offs

- [Migration on large databases] → Task tables are small in a personal GTD app; table recreation is fine.
- [Missed references] → Thorough grep + test suite catches any remaining `pending` strings.
- [Assignee nil checks] → Compiler catches all `string` → `*string` type mismatches; no silent runtime failures.
- [Kind data loss] → The `kind` column data is discarded. Acceptable: `kind` was always derivable from `assignee`, and the column held no unique information.