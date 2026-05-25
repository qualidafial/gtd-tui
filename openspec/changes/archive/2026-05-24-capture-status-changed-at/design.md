## Context

Status transitions flow tasklist → `taskstatus` overlay (a `huh.Confirm`) → `apply(svc, ctx, id)` → `TaskService.CompleteTask/DropTask/ReopenTask(ctx, id)` → `sqlite.transitionTask`, which sets `status=?`, `updated_at=now()`, and rewrites `order_key`. The `tasks` table tracks `created_at` and `updated_at` only; `updated_at` is overwritten by every edit, so the moment of a status change is not durably recorded. A reusable date input already exists at `tui/components/date.Field` (natural-language parsing, date-only vs timed, `*time.Time` accessor) — the same component the edit form uses.

## Goals / Non-Goals

**Goals:**
- A durable event-time field for the last status transition, distinct from record-time `updated_at`.
- Let the user assert the true transition time (backdate) from the confirmation overlay, defaulting to now.
- Keep the fast path (accept now, confirm) at Enter-Enter.

**Non-Goals:**
- No transition/activity history table — a single overwritten field only (history lands with `implement-timelines`).
- No per-status columns (`CompletedAt`/`DroppedAt`) — one status-neutral field.
- No display of `StatusChangedAt` as a chip in the task *list* yet (the editor header is in scope; the list chip is not).
- No `ctrl+enter` submit-from-any-field hotkey.
- No redesign of the editor — `task-edit-ui` documents existing behavior; only the Status line is new.

## Decisions

**Single status-neutral `StatusChangedAt`, overwritten per transition.** Set to `CreatedAt` at creation (creation is the transition into `pending`), then to the supplied instant on every Complete/Drop/Reopen. Answers "current status since when." A reopen overwrites a prior completion time — acceptable; durable history is explicitly deferred. Status-neutral name parallels `created_at`/`updated_at` and avoids column sprawl.

**Event time vs record time are both written on a transition.** `updated_at` stays record time (always `now()`); `status_changed_at` is the user-supplied event time. A backdated completion therefore still advances `updated_at`. This is intentional: a backdated close sorts to the top of the `updated_at DESC` fallback ordering, but closed tasks have a null `order_key` and this is only the closed-list tiebreak, so the effect is cosmetic.

**Widen the three transition signatures with `at time.Time`.** `CompleteTask`/`DropTask`/`ReopenTask(ctx, id, at)` across the `gtd.TaskService` interface, `service`, and `sqlite.transitionTask`. `CreateTask` keeps its signature and sets `StatusChangedAt = CreatedAt` internally. (A later `comment` parameter from `implement-comments` will widen these again — anticipated, not handled here.)

**Empty timestamp falls back to `now()`.** The overlay's date field is prefilled with the current local time; if the user clears it, `apply` substitutes `now()` rather than erroring, so the field is never null and the Enter-Enter path needs no special value. No validation gates on the instant (future or before-`created_at` allowed) — backdating to assert truth is the feature.

**Migration `0002` rebuilds the table to add the column.** SQLite forbids a non-constant default on `ADD COLUMN` against a non-empty table, so the additive `ALTER TABLE ... ADD COLUMN ... DEFAULT (strftime(...))` fails for any user who already has tasks. Instead, rebuild: `CREATE TABLE tasks_new` carrying `status_changed_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))` (a fresh CREATE allows the non-constant default that ADD COLUMN does not), `INSERT ... SELECT` copying each row's `updated_at` into `status_changed_at` (the backfill — closest available proxy for the last transition), `DROP TABLE tasks`, rename `tasks_new` to `tasks`, recreate the index. No other table references `tasks`, so no foreign-key juggling is needed.

**Overlay becomes a two-field `huh.Form`.** A `date.Field` (bound to a `*time.Time` prefilled with now) followed by the existing confirm. The affirmative-preselect and esc-cancel behavior carry over. `apply` reads both the confirm bool and the chosen instant.

**Extract the WHEN formatter to `internal/reltime`.** `formatWhen` and its helpers (`truncateToDay`, `isLocalMidnight`, `formatClock`) are currently private to `tui/pages/tasks/tasklist`. The editor's Status line needs the same past-ladder vocabulary, so move them verbatim into a new `internal/reltime` package (exported `reltime.Format(ref, now)`), and have `tasklist` call it. No behavior change — the existing render tests continue to pin the ladder (relocated with the code). This keeps one relative-time vocabulary across the list and the editor.

**Status line uses the past ladder against `StatusChangedAt`.** The editor renders `Status: <Title-cased status> (<reltime.Format(StatusChangedAt, now)>)`. Status changes are always in the past (or now), so the past ladder applies: `today` / `Nd` (≤30) / absolute date, with a same-day change showing the clock time. Title-casing maps the lowercase `TaskStatus` constant to a display form (e.g. `pending` → `Pending`). Consistency with the list's chip vocabulary is preferred over a bespoke "duration" rendering.

**Document the editor as `task-edit-ui` (new capability).** The editor has no spec today; rather than spec only the new line in isolation, capture the editor's existing behavior (fields, defaults, header, save/create-update, error retry, back-out) as ADDED requirements in a new capability, with the Status line as one of them. This records reality once and gives the new behavior a home.

## Risks / Trade-offs

- [Backfill is imprecise for rows whose `updated_at` reflects a non-status edit] → Accepted; `updated_at` is the best proxy available for pre-existing data and this is a personal tool with little history.
- [Signature churn on the transition methods, twice (now + future comments)] → Accepted; the user has confirmed both passes are expected.
- [Reopen loses the prior completion instant] → Accepted; single-field simplicity is preferred until the timelines/activity-log work.
- [Date field cleared by the user] → Handled by the `now()` fallback in `apply`; never persists null.

## Open Questions

None outstanding — forks resolved during exploration (single field, status-neutral, signature widening accepted, Enter-Enter UX, additive+backfill migration).
</content>
