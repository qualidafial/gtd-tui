## Context

Setting Due / Defer Until today requires the full task editor (`e`), whose form walks title → description → assignee → due → defer → status. There is no direct path to scheduling. Meanwhile, status transitions already have a well-worn pattern: a dedicated key (`space`/`delete`) pushes a tiny single-field overlay (`tui/pages/tasks/taskstatus`) that edits one value on a `datefield` and commits through a service call.

The pieces this change needs already exist:
- `tui/components/form/datefield` — a `*time.Time` field with natural-language parsing, empty→nil, and absolute-date echo on blur.
- `TaskService.UpdateTask` — persists `Task.Due` and `Task.DeferUntil`.
- The `screen.Push(overlay)` dispatch used by every list/view action.

`taskstatus` is the near-exact template: `form.New(datefield, savefield)`, an `applying` guard, and a commit command. The only differences for date editing are which struct field is written and that there is no separate "confirmation instant" — the date *is* the value.

## Goals / Non-Goals

**Goals:**
- One direct keystroke to set/change/clear each of Due (`d`) and Defer (`f`) on the focused task.
- A single overlay serving both dates, parameterized by target field.
- Zero domain/service change; reuse `datefield` and `UpdateTask` as-is.

**Non-Goals:**
- Any project-side date editing (projects have no Due/Defer fields).
- Relative quick-set chords (`d` then `t` = today); `datefield`'s natural-language input already covers "today"/"tomorrow"/"next friday".
- Defer/Due cross-validation (e.g. warn when defer ≥ due) — deferred; the fields are independent for now.
- Multi-select / bulk date edits.

## Decisions

### Decision: One overlay package `taskdate` parameterized by target field
Create `tui/pages/tasks/taskdate/` modeled on `taskstatus`: `form.New(dateField, saveField)` with an `applying` guard and a commit command. A `Field` value (`Due` | `Defer`) selects the label, which struct member is prefilled, and which member is written back before `UpdateTask`.

- **Why:** Due and Defer differ only by field + label. A parameterized overlay mirrors how `taskstatus` is parameterized by `Transition`, keeping one code path and one set of tests.
- **Alternatives considered:** (a) Two separate packages — rejected as near-duplicate code. (b) Generalize `taskstatus` itself to cover dates too — rejected: status carries a confirmation-instant + transition-to-service mapping that date editing does not need; conflating them muddies both.

### Decision: Prefill current value; empty submit clears
The field opens seeded with the task's current target date (or empty if unset). Because `datefield` maps empty input to `nil`, submitting an empty field clears the date. This makes `d`/`f` a single set/change/clear affordance with no separate "clear" key.

- **Why:** Matches how the full editor already round-trips these fields, and keeps the surface to two keys total.
- **Alternatives considered:** Default the field to `now` like `taskstatus` — rejected: `taskstatus` needs an instant for the status change; a date editor should show the value being edited, and defaulting to today would silently propose a due date the user didn't ask for.

### Decision: Commit writes only the target field via `UpdateTask`
The commit command copies the loaded task, overwrites just `Due` or `DeferUntil` with the parsed value (possibly nil), and calls `UpdateTask`, preserving every other attribute — the same read-modify-write `taskedit` performs.

- **Why:** No new service method; `UpdateTask` already persists both fields. Touching only the target member avoids clobbering concurrent edits to other attributes captured in the loaded task.

### Decision: Bindings live on the task list and task view; project view inherits for free
Add `SetDue`/`SetDefer` bindings to `tasklist.KeyMap` and `taskview.KeyMap`, each with a dispatch case that pushes `taskdate.New(task, svc, field)`. The project view embeds the task list, so its selected-row `d`/`f` fall through to the embedded list with no project-view change.

- **Why:** Due/Defer are Task-only; these are the two screens where a task is focused. Reuses the established `screen.Push` action pattern.

## Risks / Trade-offs

- **Help-bar crowding** on the task list (already dense) → two more entries. Accepted; group `d`/`f` with the other task-action bindings.
- **`f` collides with a future Linear-style `f`=filter** if that direction is ever revived → currently filter is `/` and `f` is free; note the contest but do not pre-solve it.
- **No defer/due sanity check** (defer after due) → out of scope; revisit if it proves confusing in use.

## Open Questions

- None blocking. Cross-field validation and relative quick-set chords are explicitly deferred (see Non-Goals).
