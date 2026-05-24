## Why

The task list shows only the title, so status, due/defer dates, and assignee are invisible until you open a task. The list can't surface urgency or the GTD signals (deferred-and-now-ready, overdue, delegated) that drive what to act on next.

## What Changes

- Replace the `list.NewDefaultDelegate()` in the task list with a custom delegate that renders each task as a status marker, the title, and inline data chips.
- Add status markers: `[ ]` pending, `[x]` done (dim green title), `[-]` dropped (dim gray, strikethrough title).
- Add relative-time data chips with four words sharing one WHEN ladder:
  - `due:` / `overdue:` for the due date (flips at **end of day** for date-only values).
  - `defer:` / `ready:` for the defer date (flips at **start of day** for date-only values); `ready:` marks a deferred task that has resurfaced.
  - WHEN renders as today/tomorrow/weekday/`Nd` up to 30 days, and an absolute `YYYY-MM-DD` date beyond 30 days.
- Add an `@assignee` chip for delegated tasks.
- Colorize chips by urgency (overdue red → due-today orange → soon yellow → later dim; defer dim blue; ready teal; assignee magenta), using huh theme colors where available.
- Suppress `due:`/`defer:` chips on done & dropped tasks; keep `@assignee` on done tasks.
- Selection highlight affects the title only; chips keep their own colors on the selected row.
- **BREAKING** (boundary semantics): date-only `defer:` query predicates now threshold at **start of local day** instead of end of day, so the live filter agrees with the displayed chip on the boundary day. `due:` stays end-of-day.

## Capabilities

### New Capabilities
- `task-list-presentation`: how a task renders in the list — status marker, title styling per status, relative-time chip vocabulary and the WHEN ladder, chip ordering/suppression rules, and urgency color bands.

### Modified Capabilities
- `task-query`: date-only `defer:` predicates resolve to start-of-local-day rather than end-of-day; `due:` and other date predicates are unchanged.

## Impact

- `tui/pages/tasks/tasklist/item.go`, `tui/pages/tasks/tasklist/model.go` — custom delegate replaces the default; chip/relative-time formatting and color logic.
- `internal/taskquery/taskquery.go` — `resolveTime` gains start-of-day resolution for date-only `defer:` values.
- Background-color detection (`tea.BackgroundColorMsg`, as in `tui/components/date/date.go`) plumbed to the delegate for adaptive colors.
- No domain, service, or storage changes; `Task` struct unchanged.
