## Context

`tui/pages/tasks/tasklist.Model` already takes a `ProjectNameFunc` (`func(id int64) string`) so the per-task editor can show a "Project: <title>" meta line. The renderer (`delegate.Render` in `tui/pages/tasks/tasklist/render.go`) ignores it. Two call sites construct a tasklist:

- `tui/app.go` — global tasks tab, passes a real `projectNameFn` that calls `projectSvc.GetProject` on demand.
- `tui/pages/projects/projectview/model.go` — in-project task list, passes a closure that returns the project's title for any id.

We want the chip in the first, suppressed in the second.

## Goals / Non-Goals

**Goals:**
- One chip per row when the task has a project and the chip is enabled in this list.
- Suppressed in the project view's task list with no visual leftover (no empty space, no zero-width gap).
- Reuse the existing chip system (color, ordering, alignment).
- Avoid any extra service call beyond what's already loaded.

**Non-Goals:**
- Caching or batch-resolution. Per-row resolution is fine — the visible list is short and the resolver returning "" on error is acceptable behavior already.
- Letting the user toggle the chip at runtime.
- Showing project chips on tasks the user is editing (the editor already has its own "Project" meta line).

## Decisions

### Opt-in via constructor option, not by ProjectNameFunc presence

Reusing `ProjectNameFunc != nil` as the "render chip" signal is tempting but conflates two distinct concerns: editor lookup vs row decoration. The project view legitimately wants the editor's Project line (it tells the user which project the editor is for) but not the row decoration.

Decision: add a separate boolean (or option func) to `tasklist.New`. The chosen shape is a positional `showProjectChip bool` parameter — small surface, no option pattern overhead for a one-flag situation. The project view passes `false`; `app.go` passes `true`.

### Chip placement and color

Render the project chip in the existing chip slot, ordered **last** (after `@assignee`). Placing it last keeps the date/availability cluster intact at the leftmost chip position where the user's eye is already trained.

Color: green (`lipgloss.Color("36")` — teal-leaning green) so it is visually distinct from due-red/orange/yellow, defer-blue, ready-teal, assignee-magenta. Final palette decision happens in implementation; the spec only requires "visually distinct from the existing chips".

Format: `+<project title>` with a leading `+`, mirroring the `@assignee` convention. Truncation: when total width pressure forces truncation, the title still truncates first (existing behavior); the project chip is short relative to titles in practice. If a project title itself is very long, it gets the same `…` ANSI-truncate treatment as any chip would — but rather than add per-chip truncation, accept that long project titles consume chip-row space and the title shrinks accordingly. Practical project titles are short (2–4 words).

### Suppression rules

The project chip is suppressed on **dropped** tasks (same as all other chips). On **done** tasks it SHALL still render — knowing which project a completed item belonged to is useful in review filters. This matches the `@assignee` rule, not the date-chip rule.

### Why not cache the resolver

The resolver is invoked once per visible row per render. With ~50 visible rows and the project table on a local SQLite db, this is negligible. Adding a cache adds invalidation complexity for negligible gain.

## Risks / Trade-offs

- [Per-row `GetProject` call adds latency] → Negligible at typical list sizes; mitigated if needed by a small memoizing wrapper in `app.go`, outside this change.
- [Project title length pushing chips off-screen] → Existing title truncation absorbs the pressure; in pathological cases the title shrinks to 1 char + "…". Documented in spec.
- [Visual noise on lists where most tasks share one project] → User can filter out the project chip by viewing inside the project; otherwise mixed projects is the expected use case.