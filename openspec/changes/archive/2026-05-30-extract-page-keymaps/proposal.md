## Why

The help bar's contents were threaded through ad-hoc plumbing: `Screen.KeyMap()` returned a freshly-constructed `help.KeyMap`, and each page's private `keyMap` struct carried embedded `nav list.KeyMap` and `editing help.KeyMap` slots that callers had to populate at render time. Bindings were not callable from outside their own model (private `keys` field), so tests reached into internals via re-declared types and the action labels drifted (`Toggle` for both complete and reopen, `Enter` for "view"). The model also made it awkward to compose help across overlays: every screen that owned a child had to inspect and re-wrap its child's `KeyMap()` return value.

## What Changes

- `screen.Screen` now embeds `help.KeyMap` directly. Every screen exposes its bindings by implementing `ShortHelp()` / `FullHelp()` rather than returning a constructed object.
- Each component and page that owns key bindings gains its own `keymap.go` defining an exported `KeyMap` struct that itself implements `help.KeyMap`. `Model.KeyMap` becomes an exported field so callers and tests can reach the bindings directly.
- Page keymaps drop their embedded `nav list.KeyMap` and `editing help.KeyMap` plumbing fields. Each `Model.ShortHelp/FullHelp` now composes its own `KeyMap` with the child screens / delegates it owns (via `slices.Concat`).
- Add `tui/keymap.go` for app-wide bindings (`Quit`, `Help`). Add `keymap.go` for `tui/components/screen` (overlay `Back`), `tui/components/tabcontainer` (`Next`, `Prev`), `tui/components/querybar` (already had `KeyMap`, moves to its own file), and `tui/pages/projects/projectview` (`Edit`).
- Rename action bindings for clarity: `Toggle` → `ToggleComplete` (the label flips between "complete"/"reopen" via `SetHelp`, but the binding *name* now reflects its primary action); `projects.Enter` → `projects.View`.
- Rename `tui/components/tabcontainer/tabcontainer.go` → `model.go` to match the page/model convention.
- Delete `tui/pages/tasks/taskedit/keymap.go` and `tui/pages/tasks/taskstatus/keymap.go`. Both are huh-form-based screens; their help is supplied by huh itself, and the standalone keymap files only existed to satisfy the old `Screen.KeyMap()` shape.
- Also remove the ad-hoc adapter types (`keyMap`, `formKeyMap`, `emptyKeyMap`) declared inline in `projectedit/model.go`, `projectpicker/model.go`, `projectstatus/model.go`, `taskedit/model.go`, and `taskstatus/model.go`. They wrapped the form's `KeyBinds()` to satisfy the old `KeyMap() help.KeyMap` shape and are no longer needed once `ShortHelp`/`FullHelp` live directly on the model.
- `tui/components/screen/overlay.go`: the package-level `keyEsc` var moves into the new `screen.KeyMap` (`Back` binding) carried on the overlay struct, so the esc binding is configurable per overlay instance.

This is a **BREAKING** refactor of the Screen interface and several model types. All in-tree callers are updated in the same change.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tui-view-stack`: `Screen` interface embeds `help.KeyMap`; overlay help composition changes from `KeyMap() help.KeyMap` to direct `ShortHelp/FullHelp`.
- `tui-application`: app exposes a `KeyMap` field; help rendering composes via embedded `help.KeyMap` interface; quit/help bindings live in `tui/keymap.go`.
- `task-status-ui`: `Toggle` binding renamed to `ToggleComplete` (key and per-status label behavior unchanged).
- `project-list-ui`: `Toggle` → `ToggleComplete`; `Enter` (enter key for "view") → `View` (same key, clearer name).

## Impact

- Code (staged):
  - `tui/components/screen/{screen.go,overlay.go,keymap.go}` — Screen embeds help.KeyMap; overlay refactor; new keymap.go.
  - `tui/components/tabcontainer/{model.go (renamed),keymap.go}` — split keymap, file rename.
  - `tui/components/querybar/{querybar.go,keymap.go}` — keymap moves to its own file.
  - `tui/{app.go,keymap.go,app_test.go}` — app KeyMap and help composition.
  - `tui/pages/projects/{keymap.go,projectlist.go,model_test.go,render.go}` — exported KeyMap, renamed Toggle/Enter, help composition.
  - `tui/pages/projects/projectedit/model.go`, `projectpicker/model.go`, `projectstatus/model.go`, `projectview/{model.go,keymap.go}` — model-level help composition.
  - `tui/pages/tasks/tasklist/{keymap.go,model.go,model_test.go,render.go}` — same shape; Toggle → ToggleComplete.
  - `tui/pages/tasks/taskedit/{model.go}` (keymap.go deleted), `tui/pages/tasks/taskstatus/{model.go}` (keymap.go deleted) — huh-driven help.
- No service / sqlite / domain changes. No DB migration.
- Visible behavior: the help bar's set of bindings is unchanged at the user level, but the grouping/ordering shifts slightly. Previously tasklist/projectlist emitted explicit nav-key groups built from the underlying `list.KeyMap`; now nav bindings arrive as part of the inner list's own help and the page only contributes its own action bindings. The total binding set and key labels are the same.
- API breaks visible to callers/tests:
  - `Screen.KeyMap() help.KeyMap` removed; `Screen` embeds `help.KeyMap` directly.
  - `m.keys` (private) replaced by `m.KeyMap` (exported) on every page Model that exposes one.
  - Renames `Toggle → ToggleComplete` and `projects.Enter → projects.View` ripple through every test that references those binding names.
  - `app_test.go` test stub `stubScreen` updated to implement `ShortHelp/FullHelp` instead of `KeyMap()`.