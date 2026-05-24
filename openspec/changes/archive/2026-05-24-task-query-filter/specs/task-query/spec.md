## ADDED Requirements

### Requirement: Query string tokenization
The system SHALL provide a parser in `internal/taskquery` exposing `Parse(query string) (gtd.TaskFilter, error)`. The query SHALL be split on whitespace into tokens. A token of the form `key:value` whose key is a recognized key SHALL set a structured filter field; any other token (including `key:value` with an unrecognized key) SHALL be treated as a free-text term.

#### Scenario: Recognized key sets structured filter
- **WHEN** parsing `status:done`
- **THEN** the resulting TaskFilter has Status = TaskStatusDone

#### Scenario: Unrecognized key becomes free text
- **WHEN** parsing `foo:bar`
- **THEN** `foo:bar` is treated as a free-text term, not an error

#### Scenario: Bare word is free text
- **WHEN** parsing `bob`
- **THEN** `bob` is added as a free-text term

### Requirement: Recognized keys
The parser SHALL recognize the keys `status`, `kind`, `assignee`, `due`, `defer`, and `ready`. `status` accepts pending/done/dropped; `kind` accepts next_action/delegated; `assignee` accepts any string; `due`, `defer`, and `ready` accept date-predicate values (`ready` accepts only threshold values, not `none`/`any`).

#### Scenario: status key
- **WHEN** parsing `status:dropped`
- **THEN** TaskFilter.Status = TaskStatusDropped

#### Scenario: kind key
- **WHEN** parsing `kind:delegated`
- **THEN** TaskFilter.Kind = TaskKindDelegated

#### Scenario: assignee key
- **WHEN** parsing `assignee:bob`
- **THEN** TaskFilter.Assignee = "bob"

#### Scenario: ready key
- **WHEN** parsing `ready:now`
- **THEN** TaskFilter.Ready is an AvailableAsOf predicate resolved to the current instant

### Requirement: Single value per key, last wins
When a recognized key appears more than once, the last occurrence SHALL win.

#### Scenario: Repeated key
- **WHEN** parsing `status:done status:dropped`
- **THEN** TaskFilter.Status = TaskStatusDropped

### Requirement: Free-text term semantics
Free-text terms SHALL be collected into TaskFilter.Search ([]string). Multiple terms SHALL be ANDed: a task matches only if every term matches. A single term matches if it appears as a case-insensitive substring of the task's title, description, OR assignee.

#### Scenario: Multiple terms are ANDed
- **WHEN** parsing `report bob`
- **THEN** TaskFilter.Search contains both "report" and "bob"
- **AND** only tasks matching both terms (each in title, description, or assignee) are intended to match

### Requirement: Date-predicate values
`due`, `defer`, and `ready` values SHALL resolve to a reference time, or (for `due`/`defer`) the null/not-null variants. Accepted forms:

- `now`: the current instant (NOT rounded to a day boundary)
- relative duration: `-Nd`, `Nd`, `Nw` (units `d`=days, `w`=weeks; leading `-` = past, otherwise future, relative to today)
- keyword alias: `overdue` ≡ `-1d`, `today` ≡ `0d`, `week` ≡ `7d`
- explicit ISO date: `YYYY-MM-DD`
- `none` / `any`: the null / not-null variants (accepted by `due` and `defer`, not `ready`)

Except for `now`, a value SHALL resolve to the end-of-day of the named date in the local timezone; `now` resolves to the current instant. A `DatePredicate` value on TaskFilter SHALL carry the resolved time and a kind discriminator (see "Date-predicate kinds and direction").

#### Scenario: now resolves to the current instant
- **WHEN** parsing `ready:now`
- **THEN** TaskFilter.Ready resolves to the current instant, not end-of-day today

#### Scenario: Relative duration in the future
- **WHEN** parsing `due:7d`
- **THEN** TaskFilter.Due is a threshold predicate resolved to end-of-day 7 days from today

#### Scenario: Relative duration in the past
- **WHEN** parsing `due:-5d`
- **THEN** TaskFilter.Due is a threshold predicate resolved to end-of-day 5 days ago

#### Scenario: Week unit
- **WHEN** parsing `defer:2w`
- **THEN** TaskFilter.Defer is a threshold predicate resolved to end-of-day 14 days from today

#### Scenario: Keyword aliases
- **WHEN** parsing `due:overdue`, `due:today`, or `due:week`
- **THEN** they resolve identically to `due:-1d`, `due:0d`, and `due:7d` respectively

#### Scenario: Explicit ISO date
- **WHEN** parsing `due:2026-06-01`
- **THEN** TaskFilter.Due is a threshold predicate resolved to end-of-day 2026-06-01 (local)

#### Scenario: None and any variants
- **WHEN** parsing `defer:none` or `defer:any`
- **THEN** TaskFilter.Defer is the IsNull or IsNotNull variant respectively

### Requirement: Date-predicate kinds and direction
A DatePredicate kind SHALL be one of: `OnOrBefore` (for `due`), `AvailableAsOf` (for `ready`), `After` (for `defer`), `IsNull` (`none`), `IsNotNull` (`any`). The selection semantics SHALL be:

- `due:X` (OnOrBefore) → `due ≤ X` — cumulative "must-deliver as of X"; excludes NULL due.
- `ready:X` (AvailableAsOf) → `defer_until IS NULL OR defer_until ≤ X` — "available as of X"; includes never-deferred tasks.
- `defer:X` (After) → `defer_until > X` — "still parked as of X"; excludes NULL defer_until.
- `none` (IsNull) → column IS NULL; `any` (IsNotNull) → column IS NOT NULL.

#### Scenario: due is cumulative upper bound
- **WHEN** filtering with `due:today`
- **THEN** tasks due today AND tasks already overdue are selected, and tasks with no due date are not

#### Scenario: ready includes never-deferred and opened gates
- **WHEN** filtering with `ready:now`
- **THEN** tasks with NULL defer_until OR defer_until ≤ now are selected (the available-now set)

#### Scenario: defer is strict lower bound
- **WHEN** filtering with `defer:2d`
- **THEN** only tasks whose defer_until is after end-of-day +2 are selected (still parked then)

### Requirement: Parse error reporting
A recognized key with an invalid value SHALL produce an error. An unrecognized key SHALL NOT produce an error (it is free text). A parse error SHALL carry the `[start, end)` rune-offset range of the offending substring within the input query, in addition to a human-readable message, so a caller can highlight the offending section.

#### Scenario: Invalid enum value with range
- **WHEN** parsing `status:bogus`
- **THEN** Parse returns an error whose message identifies `status:bogus`
- **AND** the error's range covers the `status:bogus` token's offsets in the input

#### Scenario: Invalid date value with range
- **WHEN** parsing `kind:delegated due:notadate`
- **THEN** Parse returns an error covering the offsets of the `due:notadate` token (not the whole input)

#### Scenario: Empty query
- **WHEN** parsing an empty string
- **THEN** Parse returns a zero-value TaskFilter and no error
