## Why

GTD practice often surfaces work that is *already finished* by the time you record it — a "I just did this" memorial entry captured for the record, momentum, or project history. Today the task editor forces every new task to `open`, so the only way to log a completed task is create-then-complete: two steps and an open task that briefly clutters the active list.

## What Changes

- The new-task editor presents a terminal **Status** choice (`Open` / `Done`, default `Open`) in place of the plain `[ Create ]` button. Selecting a status and pressing Enter creates the task directly in that status — one step, no extra keypresses on the default (`Open`) path.
- Creating a task with `Done` status writes it straight to done: no order key, `StatusChangedAt` stamped at creation. The persistence layer already supports this; only the entry point is new.
- Editing an existing task is unchanged: status remains non-editable in the editor and continues to flow through the dedicated transition overlay. The existing-task form keeps its plain `[ Save ]` terminal button.
- The form framework gains a general rule: **Enter on the last visible field submits** when that field does not itself claim Enter. This lets the status radio act as the terminal submit affordance without bespoke wiring, and reframes the save button as a focus placeholder with no behavior of its own.
- With the form owning Enter-at-the-end, the `selectfield.WithSubmitOnEnter` mode is removed: a `selectfield` now claims Enter only while filtering (to accept the filter) and otherwise defers to the form's last-field rule. The project and task pickers drop the option.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `task-edit-ui`: new-task form offers an `Open`/`Done` status choice and creates into the chosen status; existing-task status editing stays out of scope.
- `form-field-toolkit`: the form submits on Enter at the last visible field when the focused field does not claim Enter; the save button becomes a valueless focus placeholder.

## Impact

- **Code**: `tui/components/form/form.go` (Enter-at-last-field submit), `tui/components/form/savefield` (drop its Enter-claim / submit emission), `tui/components/form/selectfield` (remove `WithSubmitOnEnter`; claim Enter only while filtering), `tui/pages/projects/projectpicker` and `tui/pages/projects/taskpicker` (drop the option), `tui/pages/tasks/taskedit/model.go` (status radio on create, `saveCmd` create branch reads the chosen status).
- **No changes** to domain, sqlite, or service layers — `sqlite.CreateTask` already handles closed-status creation (skips `order_key`, sets `StatusChangedAt = now`).
- **No new dependencies.** Reuses the existing `radiofield` component.
- **Forward note (not in scope):** if status editing is later added to the existing-task form, the read-only `Status: <Status> (<WHEN>)` header line should become conditionally visible based on whether status changed in that edit session.
