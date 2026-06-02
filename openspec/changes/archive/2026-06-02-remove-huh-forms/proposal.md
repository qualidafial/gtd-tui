## Why

`huh.Form` orchestration is the wrong shape for our overlays. Three concrete pains:

1. **Save latency.** A submit-from-any-field shortcut has to walk fields by feeding messages into `huh.Form.Update`, which triggers focus/blur cmds, blink timers, `UpdateFieldPositions`, and full viewport rebuilds per step — visibly slow on a 5-field overlay (see superseded `submit-form-shortcut`).
2. **Field-position coupling.** `Enter`'s meaning depends on which field is focused (`Next` on non-last, `Submit` on last). That's surprising for users and forces every cross-cutting concern (ctrl+s, validation walking) through huh's message pipeline.
3. **Hidden state.** Public API exposes `Form.State` but the transitions that drive it (`nextFieldMsg`, `nextGroupMsg`, group selector navigation) are internal. Each overlay carries the same "is the form completed/aborted/normal" boilerplate, and our `formx` package had to poke private message types to make ctrl+s work.

We aren't using anything `huh` gives us that we can't build in less code with `bubbles/textinput`, `bubbles/textarea`, and explicit focus management. Replacing `huh.Form` lets save be a synchronous function call, makes keybindings live in one place per screen, and removes a dependency.

## What Changes

- **BREAKING (internal):** Remove `charm.land/huh/v2` from the dependency graph.
- Introduce a small `tui/components/formfield` package (working name) that owns:
  - A `Field` interface (`Focus`, `Blur`, `View`, `Update`, `Validate`, `Error`) backed by thin wrappers around `bubbles/textinput`, `bubbles/textarea`, plus a yes/no confirm primitive and a single-select primitive.
  - A `Form` helper that holds `[]Field`, current focus index, and synchronous `Submit() error` / `Next()` / `Prev()` methods. No internal message pipeline — overlays drive it directly from their `Update`.
- Migrate every overlay off `huh.Form`/`huh.Group`/`huh.Field`:
  - `itemcapture` (2 fields)
  - `taskedit` (5 fields incl. date)
  - `projectedit` (4 fields incl. date)
  - `taskstatus` (date + confirm)
  - `projectstatus` (date + confirm)
  - `projectpicker` (single select)
- Reimplement `tui/components/date.Field` against the new `Field` interface (drop `huh.Field` / `huh.Accessor` / `huh.Theme` / `huh.InputKeyMap` dependencies).
- Delete `tui/components/formx` (the formx workaround built atop `huh.Form` is no longer needed).
- `Ctrl+S` save-from-anywhere is preserved and becomes a one-line synchronous call.

## Capabilities

### New Capabilities
- `form-field-toolkit`: The shared field/form primitives that replace `huh.Form`/`huh.Group` — keymap, focus model, submit semantics, validation contract, and the wrapped bubbles widgets.

### Modified Capabilities
None. The user-visible behavior of `task-edit-ui`, `project-edit-ui`, `inbox-page`, `task-status-ui`, `project-list-ui`, and `project-picker-overlay` does not change at the spec level — only the implementation underneath. The `submit-form-shortcut` change (which introduced `formx`) is superseded by this one; its spec was never archived into `openspec/specs/`, so there is nothing to delta. The `ctrl+s` save-from-anywhere contract moves into `form-field-toolkit`.

## Impact

- Code: every overlay listed above, plus `tui/components/date/`, plus the new `tui/components/formfield/`, plus deletion of `tui/components/formx/`.
- Dependencies: drop `charm.land/huh/v2`. Already on `bubbles/v2` and `bubbletea/v2`.
- Risk: large blast radius (six overlays + one shared component) but each migration is independent and testable. The existing real-stack overlay tests are the safety net — they exercise the user-visible contract, not the huh-specific internals.
- UX: should be invisible to users except that save is fast. Keybindings, field order, validation messages all preserved.
- Performance: the explicit motivator. Ctrl+S becomes O(fields) synchronous validation; no message-loop overhead.
