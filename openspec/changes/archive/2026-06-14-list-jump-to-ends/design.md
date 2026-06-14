## Context

Reordering for both tasks and projects is implemented with fractional-indexed `order_key` strings (the `orderkey` package). The SQLite layer's `shiftTask` (and the project equivalent) computes the moving item's current position `pos` within the filtered, same-status set, derives a target `newPos = pos + delta` for a ±1 move, looks up the prev/next neighbor keys at that target, and slots a new key via `orderkey.Between`. On `Between` exhaustion it renumbers the whole same-status set. The TUI layers (`tasklist`, `projectlist`) expose this through `moveCmd(direction int)` bound to `shift+↑`/`shift+↓`, with per-selection enablement in `updateKeybindings`.

Move-to-top and move-to-bottom are the same operation with an absolute target position (`0` and `len(others)`) instead of a relative delta. The existing prev/next-key + `Between` + renumber-fallback machinery already handles arbitrary target positions; only the position *derivation* differs.

## Goals / Non-Goals

**Goals:**
- Add `shift+home` (move first) / `shift+end` (move last) on the task list and project list.
- Reuse the existing filter-scoping, closed-item rejection, and order-key exhaustion fallback unchanged.
- Keep the new service methods symmetric with `MoveTaskUp`/`Down` and `MoveProjectUp`/`Down` in signature and semantics.

**Non-Goals:**
- No new keys for cursor navigation; `home`/`end` (GoToStart/GoToEnd) already jump the cursor without reordering and are untouched.
- No schema or migration changes.
- No reorder support on lists that don't already have it (inbox, timeline, notes).

## Decisions

**Generalize the SQLite reorder helper to an absolute target position.**
Extract the position-application tail of `shiftTask` into `moveTaskTo(ctx, id, newPos, filter)` (and the project equivalent). `shiftTask` keeps computing `pos` and calls `moveTaskTo(ctx, id, pos+delta, filter)`; `MoveTaskFirst` calls it with `newPos = 0`, `MoveTaskLast` with `newPos = len(others)`. The neighbor lookup, `orderkey.Between`, and the renumber fallback are shared verbatim.
*Alternative considered:* a dedicated top/bottom path that grabs only the single boundary neighbor key and calls `orderkey.Between(prev, "")` / `Between("", next)`. Rejected — it duplicates the exhaustion-fallback logic and diverges from the proven `shiftTask` path for no real gain.

**Top/bottom is a no-op when already at the boundary, mirroring the one-slot guard.**
`MoveTaskFirst` on the first filtered item and `MoveTaskLast` on the last are no-ops (no key changes), exactly as `pos+delta` out of range is today. The UI disables the binding in those cases via `updateKeybindings`, so reaching the service with a no-op only happens on a race.

**UI enablement reuses the existing move-up/move-down predicates.**
`MoveFirst` is enabled exactly when `MoveUp` is (orderable selection that is not already first in the filtered set); `MoveLast` exactly when `MoveDown` is (orderable, not already last open). No new selection analysis is required — the same `idx`/status checks already computed in `updateKeybindings` drive all four bindings.

**`moveCmd` takes the target instead of only a direction.**
Extend the TUI `moveCmd` to dispatch to one of the four service methods. The post-move reload + reselect-by-id flow (`tasksReorderedMsg` / project equivalent) is unchanged; only the chosen service call differs.

## Risks / Trade-offs

- [Renumber cost on a move-to-bottom across a huge list] → Same fallback path and cost as the existing one-slot move on key exhaustion; move-to-bottom is not more expensive than a worst-case `shift+↓`. No new risk.
- [Binding discoverability — four shift+arrow/home/end bindings] → `shift+home`/`shift+end` render in the same help group as the existing move bindings and are only shown when enabled, so the help bar stays accurate per selection.

## Migration Plan

Additive only. New service methods and keybindings ship together; no data migration, no rollback concerns beyond reverting the commit.
