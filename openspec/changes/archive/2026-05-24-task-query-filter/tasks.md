## 1. Domain — TaskFilter extensions

- [x] 1.1 Add DatePredicate type to task.go (kinds: OnOrBefore, AvailableAsOf, After, IsNull, IsNotNull; time-based kinds carry a resolved time)
- [x] 1.2 Add fields to TaskFilter: Assignee *string, Due *DatePredicate, Ready *DatePredicate, Defer *DatePredicate, Search []string
- [x] 1.3 Remove IncludeDeferred from TaskFilter (delta against archived task-service); update any existing callers
- [x] 1.4 Keep existing builder methods working; add WithAssignee / WithSearch helpers if convenient

## 2. Query parser (internal/taskquery)

- [x] 2.1 Create internal/taskquery package with Parse(query string) (gtd.TaskFilter, error)
- [x] 2.2 Tokenize on whitespace; classify key:value (recognized key) vs free-text term
- [x] 2.3 Map status/kind/assignee keys to filter fields with enum validation (invalid value → error)
- [x] 2.4 Resolve date values in one helper: `now` (instant), relative durations (-Nd/Nd/Nw), keyword aliases (overdue=-1d, today=0d, week=7d), ISO date, and none/any; non-`now` values resolve to end-of-local-day
- [x] 2.5 Map due→OnOrBefore, ready→AvailableAsOf, defer→After; due/defer also accept none/any (IsNull/IsNotNull); ready accepts threshold values only
- [x] 2.6 Collect free-text terms into Search; unrecognized key:value treated as free text
- [x] 2.7 Single-value-per-key last-wins; empty query → zero filter, no error
- [x] 2.8 Define a ParseError carrying a human message plus the [start,end) rune-offset range of the offending token

## 3. Parser tests

- [x] 3.1 Table-driven tests: each recognized key (incl. ready) → expected filter field/kind
- [x] 3.2 Tests for free-text collection and AND semantics
- [x] 3.3 Tests for date values: now (instant, not end-of-day), relative, aliases, ISO, none/any
- [x] 3.4 Tests for error cases (status:bogus, kind:foo, due:notadate, ready:none) and unrecognized-key-as-text; assert the error's range covers exactly the offending token (not the whole input)
- [x] 3.5 Timezone boundary tests for now-vs-today and overdue/today/week across local offsets

## 4. SQLite — ListTasks filtering

- [x] 4.1 Apply TaskFilter.Search: per term, lower(title/description/assignee) LIKE %term% ANDed across terms
- [x] 4.2 Apply TaskFilter.Assignee: case-insensitive substring on assignee
- [x] 4.3 Apply Due (`due IS NOT NULL AND due <= t`), Ready (`defer_until IS NULL OR defer_until <= t`), Defer (`defer_until > t`); IsNull/IsNotNull → `column IS [NOT] NULL`
- [x] 4.4 Remove implicit deferral filtering from ListTasks (delta: REMOVED "Deferred task filtering in queries"); ensure clauses compose with Status/Kind

## 5. SQLite tests

- [x] 5.1 Test free-text match across title/description/assignee and multi-term AND
- [x] 5.2 Test assignee filter
- [x] 5.3 Test due (cumulative ≤, includes overdue, excludes NULL), ready (includes NULL and passed gates), defer (strict >, excludes NULL), and IsNull/IsNotNull
- [x] 5.4 Verify no implicit deferral filtering: with no Ready/Defer predicate, future-deferred tasks are returned

## 6. TUI — query bar

- [x] 6.1 Add a query-bar textinput to tui/pages/tasks/tasklist (always visible, shows active query)
- [x] 6.2 Focus on `/`; disable the list's built-in filter keybinding
- [x] 6.3 Live parse for validation: on Enter and on a 2s debounce after last keystroke; live parse updates error display only (no reload)
- [x] 6.4 Apply on Enter only: on successful parse, reload via ListTasks; on failure, do not reload
- [x] 6.5 Esc reverts to last-applied query without reloading
- [x] 6.6 On parse error, render the message inline and highlight the offending substring in the bar using the error's range; keep current results
- [x] 6.7 Seed the default `status:pending ready:now` query in tui/app.go (replace hard-coded TaskFilter)

## 7. Verification

- [x] 7.1 go build ./... and go test ./... pass
- [x] 7.2 Manually drive the TUI: default view (status:pending ready:now), status:done, defer:any (parked pile), free-text search, an invalid query
