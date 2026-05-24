## Why

The task list view shows a single hard-coded `pending` filter with no way to change it — users cannot review completed or dropped work, narrow to delegated items, or search by text. The GTD Engage/Reflect workflows depend on slicing the task list freely. A query box that accepts a compact `key:value` plus free-text syntax (e.g. `status:done kind:delegated bob`) gives one low-ceremony control for all of it.

## What Changes

- Add a query language for the task list: space-separated tokens, where `key:value` tokens map to structured filters and bare words are free-text terms.
- Supported keys (first cut): `status`, `kind`, `assignee`, `due`, `defer`, `ready`.
- Free-text terms match (case-insensitive substring) against title, description, and assignee; multiple terms are ANDed.
- Date predicates resolve to a reference time (`now`, relative `-5d`/`7d`/`2w`, keyword aliases, ISO date) or `none`/`any`; `due` filters `≤`, `ready` filters availability (`null or ≤`), `defer` filters "still parked" (`>`).
- Add a parser (new `internal/taskquery` package) that turns a query string into a `TaskFilter`, with structured parse errors for invalid keys/values.
- **BREAKING (internal API):** extend `TaskFilter` with `Search []string`, `Assignee *string`, and `Due`/`Ready`/`Defer` date-predicate fields, and **remove `IncludeDeferred`**. Deferral hiding is no longer implicit — it is expressed explicitly via `ready`/`defer` predicates.
- Extend the sqlite `ListTasks` implementation to apply the new fields (LIKE clauses for text/assignee, threshold predicates for due/ready/defer) and stop performing implicit deferral filtering.
- Add a query bar to the task list TUI: focused with `/`, editable, applies on Enter, defaults to `status:pending ready:now`, surfaces parse errors inline.

## Capabilities

### New Capabilities

- `task-query`: The query-string grammar, token semantics, date-predicate vocabulary, and the parser that converts a query string into a `TaskFilter` (including error reporting for invalid input).
- `task-list-query-ui`: The task-list query bar — focus/edit/apply/cancel interaction, default query, and inline parse-error display.

### Modified Capabilities

- `task-service`: `TaskFilter` gains `Search`, `Assignee`, and `Due`/`Ready`/`Defer` date-predicate fields and **loses `IncludeDeferred`**; `ListTasks` no longer does implicit deferral filtering.
- `task-sqlite`: `ListTasks` query construction adds case-insensitive LIKE matching (title/description/assignee) and threshold predicates for due/ready/defer; the implicit "exclude future-deferred by default" requirement is removed.

## Impact

- **New package**: `internal/taskquery` (parser + filter mapping, unit-tested independently of the TUI).
- **Root package**: `task.go` — `TaskFilter` gains fields and a `DatePredicate` type; `IncludeDeferred` is removed.
- **sqlite/**: `task.go` `ListTasks` query builder extended (LIKE + threshold predicates), implicit deferral filtering removed; new tests for text/assignee/date filtering.
- **TUI**: `tui/pages/tasks/tasklist/` gains a query bar component and wires parse → `TaskFilter` → reload; `tui/app.go` seeds the default `status:pending ready:now` query.
- **No new dependencies**: uses existing squirrel + bubbles textinput.
