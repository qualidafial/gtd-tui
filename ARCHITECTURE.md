# Architecture

## Project Layout

Follows [Ben Johnson's Go application structure](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091):

- Root package (`gtd`) — domain types and service interfaces only. No I/O, no dependencies.
- `sqlite/` — SQLite implementation of the service interfaces.
- `internal/set/` — generic Set type used internally.
- `cmd/` — entry point(s) (not yet created).

## Domain Types

Core types are plain structs with value semantics — no pointers. Defined in the root package.

**Task**: `id`, `title`, `description`, `status`, `due`, `defer_until`, `project_ids`, `note_ids`, `created_at`, `updated_at`

**Project**: `id`, `title`, `outcome`, `description`, `status`, `due`, `task_ids`, `note_ids`, `created_at`, `updated_at`

All relationships are many-to-many. Notes are not yet implemented.

## Service Interfaces

Defined in `gtd.go`. Read methods return values; write methods return the updated value with server-assigned fields (`ID`, `CreatedAt`, `UpdatedAt`) populated:

```go
CreateTask(ctx, Task) (Task, error)
UpdateTask(ctx, Task) (Task, error)
```

## SQLite Layer

**Driver**: `modernc.org/sqlite` — pure Go, no CGO.

**Query builder**: `github.com/Masterminds/squirrel` — used for all query construction. No raw string SQL outside of migrations.

**Migrations**: SQL files embedded via `//go:embed` in `sqlite/migrations/`. Named `NNNN_description.sql`, applied in lexicographic order, tracked in a `migrations` table. Each migration runs in its own transaction.

**Schema constraints**: `CHECK` constraints enforce non-empty titles and valid status enum values at the database level.

**Pragmas**: `PRAGMA journal_mode=WAL` (concurrent reads) and `PRAGMA foreign_keys=ON` (enforce `ON DELETE CASCADE`) are set on every connection.

**Transactions**: Write operations (`CreateTask`, `UpdateTask`) open a transaction at the service method level and pass it to all internal helpers. Callers see atomic commits or rollbacks.

**Relationship sync** (`syncTaskProjectIDs`): Loads the current set of IDs from the DB, diffs against the desired set, and issues only the inserts and deletes that are necessary. No delete-all + re-insert.

**Bulk ID loading**: `Tasks()` and `Projects()` load relationship IDs in a single `WHERE id IN (...)` query after fetching the primary rows, avoiding N+1 queries. Returns `map[parentID][]childID`.

## Conventions

- Value semantics throughout — no `*Task` or `*Project` in service interfaces.
- `int64` for all IDs (matches SQLite's `INTEGER PRIMARY KEY`).
- Nullable timestamps use `*time.Time`; stored as UTC.
- Use `new(expr)` (Go 1.26) to take the address of an expression result instead of helper functions like `ptrTime`.
- Tests use `github.com/stretchr/testify` (`assert` / `require`) and prefer table-driven style.
- Test databases use `:memory:` via a shared `openTestDB(t)` helper.
- File edits must use the Edit or Write tools — never `sed`/`awk` (breaks VS Code file watching).
