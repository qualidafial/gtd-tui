# Architecture

## Project Layout

Follows [Ben Johnson's Go application structure](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091):

- Root package (`gtd`) — domain types and service interfaces only. No I/O, no dependencies.
- `sqlite/` — SQLite implementation of the service interfaces.
- `service/` — cross-service orchestration (e.g. clarify flow spanning Inbox + Task + Project).
- `tui/` — Bubbletea v2 UI; pages under `tui/pages/`.
- `internal/set/` — generic Set type used internally.
- `cmd/` — entry point(s).

## Domain Types

Domain types are plain structs with value semantics. The Go source in the root package is authoritative for field lists; the semantic meaning of each entity is in `DESIGN.md`. This section documents only the relationship topology and structural choices that affect storage and querying.

**Relationship topology**:

- `Task → Project` — 0..1 via `Task.ProjectID *int64`. Standalone tasks have a nil project.
- `Item → (Task | Project | Someday | Reference)` — 0..1 via `Item.ClarifiedInto*` nullable FKs (soft-delete lineage). At most one is set; all nil means the Item is still in the inbox.
- `Comment → (Task | Project)` — exactly one of `Comment.TaskID` / `Comment.ProjectID` is set (enforced by CHECK).
- `MeetingLink (Meeting ↔ Task | Project | Item)` — m:n join row; exactly one of the target FKs is set (enforced by CHECK).
- `Project → Task` is the reverse of `Task → Project`; no separate join table.

## Service Interfaces

Defined in the root package alongside their domain types (`task.go`, `project.go`, etc.). Read methods return values; write methods return the updated value with server-assigned fields (`ID`, `CreatedAt`, `UpdatedAt`) populated:

```go
CreateTask(ctx, Task) (Task, error)
UpdateTask(ctx, Task, comment string) (Task, error)
```

Status transitions are exposed as per-transition methods (`CompleteTask`, `DropTask`, `ReopenTask`, `CompleteProject(dropOpenTasks, comment)`, etc.) rather than free-form status setters in `Update*`. This keeps cascade logic localized to one place per transition.

The clarify flow on `InboxService` (`Discard`, `Incubate`, `FileAsReference`, `ClarifyAsTask`, `ClarifyAsProject`) is transactional: each method creates the destination entity and stamps the Item's `ClarifiedInto` pointer atomically. See `DESIGN.md` for the per-terminal semantics.

## SQLite Layer

**Driver**: `modernc.org/sqlite` — pure Go, no CGO.

**Query builder**: `github.com/Masterminds/squirrel` — used for all query construction. No raw string SQL outside of migrations.

**Migrations**: SQL files embedded via `//go:embed` in `sqlite/migrations/`. Named `NNNN_description.sql`, applied in lexicographic order, tracked in a `migrations` table. Each migration runs in its own transaction.

**Schema constraints**: `CHECK` constraints enforce:
- non-empty titles where required,
- valid status / kind enum values,
- "exactly one FK set" on dual/multi-nullable-FK tables (`comments`, `meeting_links`, `items.clarified_into_*`).

**Pragmas**: `PRAGMA journal_mode=WAL` (concurrent reads) and `PRAGMA foreign_keys=ON` (enforce `ON DELETE CASCADE`) are set on every connection.

**Transactions**: Write operations open a transaction at the service method level and pass it to all internal helpers. Callers see atomic commits or rollbacks. Clarify operations and edit-with-comment operations are transactional across multiple tables.

**Relationship sync** (meeting links): Loads the current set of link rows from the DB, diffs against the desired set, and issues only the inserts and deletes that are necessary. No delete-all + re-insert.

**Bulk loading**: List endpoints load relationship rows in a single `WHERE id IN (...)` query after fetching the primary rows, avoiding N+1.

## Conventions

- Value semantics throughout — no `*Task` or `*Project` in service interfaces.
- `int64` for all IDs (matches SQLite's `INTEGER PRIMARY KEY`).
- Nullable timestamps and FKs use `*T`; timestamps stored as UTC.
- Use `new(expr)` (Go 1.26) to take the address of an expression result instead of helper functions like `ptrTime`.
- Tests use `github.com/stretchr/testify` (`assert` / `require`) and prefer table-driven style.
- Test databases use `:memory:` via a shared `openTestDB(t)` helper.
- File edits must use the Edit or Write tools — never `sed`/`awk` (breaks VS Code file watching).
