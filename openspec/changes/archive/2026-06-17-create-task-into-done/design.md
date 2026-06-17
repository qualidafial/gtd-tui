## Context

The task editor (`tui/pages/tasks/taskedit`) builds a `form.Model` whose terminal field is a `savefield` ("Create" for new tasks, "Save" for existing). `taskedit.New` force-sets `task.Status = open` for new tasks, and `saveCmd` calls `TaskService.CreateTask` or `UpdateTask`. The persistence layer already supports creating a closed task: `sqlite.CreateTask` branches on `isClosedStatus(task.Status)` — closed tasks get no `order_key`, and `StatusChangedAt` is stamped at creation alongside `CreatedAt`/`UpdatedAt`. So creating directly into done is purely an entry-point gap, not a data-model gap.

Today, Enter-submits-on-the-last-field works only because `savefield` explicitly claims Enter (via its `Keys()`) and emits `form.SubmitRequestMsg`. The form's own `Next` binding includes Enter, but `Next()` is a no-op on the last field — it does not submit. The `form.Model` already distinguishes "field claims this key" from "form navigation" via the `keymap.Handles(field, msg)` guard in `Update`.

The existing-task status-change path is the dedicated transition overlay (`task-status-ui`), and `UpdateTask` explicitly rejects status changes. This change does not touch that path.

## Goals / Non-Goals

**Goals:**
- Let a user create a task directly in `Done` ("I just did this") from the task editor in one step, with no extra keypresses on the default `Open` path.
- Offer the status choice as the terminal field so it doubles as the submit affordance.
- Generalize "Enter at the end submits" into a form-level rule, so the radio-as-terminal-field needs no bespoke submit wiring and the save button becomes a valueless focus placeholder.

**Non-Goals:**
- Editing the status of an *existing* task from the form (stays on the transition overlay; `Dropped` is not offered anywhere in the editor).
- Backdating completion. A `Done` create stamps `StatusChangedAt = now` via `CreateTask`; recording a past completion time is out of scope.
- Any domain/sqlite/service change. `CreateTask` already handles closed-status creation.

## Decisions

### Decision: Form owns "Enter on the last visible field submits"
Move submit-on-Enter-at-end from `savefield` into `form.Model`. In the `KeyPressMsg` handler, when the gesture is Enter, the focused field does not claim Enter (the existing `keymap.Handles` guard already gates this — claiming fields fall through to `field.Update`), and there is no later visible field, run `handleSubmit` instead of the no-op `Next()`. Tab/down on the last field stay no-ops (only Enter submits).

- **Why over giving `radiofield` a submit-on-enter mode (like `selectfield.WithSubmitOnEnter`):** the user's framing is that a terminal field is just a focus stop and the form owns commit-at-the-end. One rule covers every current and future terminal field instead of per-field opt-ins.
- **Audit:** every existing `form.New` call terminates in either a `savefield` or a submit-on-enter `selectfield` (the pickers) — all Enter-claiming. The new rule only fires for a *non*-claiming terminal field, which is nothing today and the new status radio tomorrow. So it is a no-op for all existing forms. `savefield` can drop its Enter claim and `SubmitRequestMsg` emission and become a pure placeholder; the form-level rule then submits when it holds focus.

### Decision: Status radio is the new-task terminal field; existing-task form unchanged
In `taskedit.New`, for a new task (`ID == 0`) replace the `savefield` with an inline `radiofield[gtd.TaskStatus]` keyed `status`, options `[Open, Done]`, default `Open` (via `radiofield.WithInitialValue(Open)`). For an existing task keep the `savefield` "Save" exactly as today, and do not force `task.Status`.

- **Why a `radiofield`:** it already exists, renders inline (`(•) Open  ( ) Done`), uses left/right to choose (no conflict with form up/down nav), and does not claim Enter — so the form-level rule submits on Enter into the highlighted option. Default `Open` ⇒ identical keypress count to today; `Done` costs one extra `→`.
- **Why not a status field for existing tasks:** out of scope (Non-Goals); the transition overlay owns that path and `UpdateTask` rejects status changes.

### Decision: Remove `selectfield.WithSubmitOnEnter`; claim Enter only while filtering
With the form owning Enter-at-the-end, the selectfield's submit-on-enter mode is redundant — except for the half it also encoded: while the list is *filtering*, Enter must accept the filter, not submit. So the mode is deleted (along with the field-level `SubmitRequestMsg` emission) and replaced by a single rule in `selectfield.Keys()`: claim Enter only when `list.FilterState() == list.Filtering`. Not filtering ⇒ Enter is unclaimed ⇒ the form's last-field rule submits a terminal selectfield; filtering ⇒ Enter is claimed ⇒ it routes to the list and accepts the filter.

- **Why now:** the form-level rule makes the mode's submit half dead weight, and a single filter-state-aware claim is simpler than a per-field opt-in flag. Both pickers (project, task) are single-field overlays whose selectfield is the terminal field, so they submit via the form rule and merely drop the option; they already react to `form.SubmittedMsg`, so no other change is needed.
- **Trade-off:** a terminal selectfield no longer advertises `enter select` in help (same cosmetic change as savefield); `ctrl+s` remains the always-visible submit affordance and Enter is a hidden alias.

### Decision: `saveCmd` reads the chosen status only on create
For a new task, read `values["status"].(gtd.TaskStatus)` and set `task.Status` before `CreateTask`. The update branch is untouched (no `status` field exists on the existing-task form, so `FieldValues()` has no `status` key). `CreateTask` does the rest (no order key, `StatusChangedAt = now`) for `Done`.

## Risks / Trade-offs

- **[A non-claiming terminal field elsewhere would newly submit on Enter]** → Audit shows none exist today (all terminal fields claim Enter). Add a form-level test for the rule so a future plain terminal field's behavior is intentional, not accidental.
- **[`savefield` losing its Enter claim changes its help/keys]** → Its `Keys()` no longer advertise "enter save"; the form's own keymap already advertises submit. Verify help rendering still shows a submit affordance on the last field.
- **[A `Done` task created with a Due/Defer set is mildly odd]** → Harmless; values persist and are ignored for closed tasks. Not worth special-casing.

## Open Questions

- None blocking. Forward note (out of scope): if existing-task status editing is later added to the form, make the read-only `Status: <Status> (<WHEN>)` header line conditionally visible based on whether status changed in that session.
