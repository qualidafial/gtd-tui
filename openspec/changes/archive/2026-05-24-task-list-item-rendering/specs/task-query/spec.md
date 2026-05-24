## MODIFIED Requirements

### Requirement: Date-predicate values
`due`, `defer`, and `ready` values SHALL resolve to a reference time, or (for `due`/`defer`) the null/not-null variants. Accepted forms:

- `now`: the current instant (NOT rounded to a day boundary)
- relative duration: `-Nd`, `Nd`, `Nw` (units `d`=days, `w`=weeks; leading `-` = past, otherwise future, relative to today)
- keyword alias: `overdue` ≡ `-1d`, `today` ≡ `0d`, `week` ≡ `7d`
- explicit ISO date: `YYYY-MM-DD`
- `none` / `any`: the null / not-null variants (accepted by `due` and `defer`, not `ready`)

Day-granularity resolution depends on which column the key thresholds, to keep the live filter consistent with the displayed chip on the boundary day:

- `due` values SHALL resolve to the **end-of-day** of the named date in the local timezone (a due date applies through the end of its day).
- `defer` and `ready` values SHALL resolve to the **start-of-day** of the named date in the local timezone (a defer date resurfaces a task at the start of its day).
- `now` SHALL resolve to the current instant regardless of key.

A `DatePredicate` value on TaskFilter SHALL carry the resolved time and a kind discriminator (see "Date-predicate kinds and direction").

#### Scenario: now resolves to the current instant
- **WHEN** parsing `ready:now`
- **THEN** TaskFilter.Ready resolves to the current instant, not a day boundary

#### Scenario: Due relative duration resolves to end of day
- **WHEN** parsing `due:7d`
- **THEN** TaskFilter.Due is a threshold predicate resolved to end-of-day 7 days from today

#### Scenario: Defer relative duration resolves to start of day
- **WHEN** parsing `defer:2w`
- **THEN** TaskFilter.Defer is a threshold predicate resolved to start-of-day 14 days from today

#### Scenario: Relative duration in the past
- **WHEN** parsing `due:-5d`
- **THEN** TaskFilter.Due is a threshold predicate resolved to end-of-day 5 days ago

#### Scenario: Keyword aliases
- **WHEN** parsing `due:overdue`, `due:today`, or `due:week`
- **THEN** they resolve identically to `due:-1d`, `due:0d`, and `due:7d` respectively (each at the resolution for its key)

#### Scenario: Explicit ISO date for due
- **WHEN** parsing `due:2026-06-01`
- **THEN** TaskFilter.Due is a threshold predicate resolved to end-of-day 2026-06-01 (local)

#### Scenario: Explicit ISO date for defer
- **WHEN** parsing `defer:2026-06-01`
- **THEN** TaskFilter.Defer is a threshold predicate resolved to start-of-day 2026-06-01 (local)

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
- **THEN** only tasks whose defer_until is after start-of-day +2 are selected (still parked then)
