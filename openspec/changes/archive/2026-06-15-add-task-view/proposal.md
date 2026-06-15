## Why

Tasks have no detail screen — pressing enter on the task list jumps straight into the editor, and there is nowhere to surface read-mostly context (the linked project today, comments and history later). Projects already have a view screen with this exact role; tasks should mirror it so the two entity types feel consistent and the planned task features have a home to land in.

## What Changes

- Add a **task view screen** that renders the task's fields (title, status, project, assignee, due, description) as a read display. For now it is a single bare panel of fields with no tab chrome; it is the future home for comments and history.
- Rebind the **task list**: enter opens the new view screen (was: edit), and `e` opens the editor — mirroring the project list. **BREAKING** for muscle memory: enter no longer opens the editor directly.
- Newly created tasks **land on their task view** after save, matching the project create-then-view flow.
- Carry the per-task action shortcuts that make sense into the view screen: `e` edit, `space` complete/reopen, `delete` drop, `p` assign-to-project, and `c` convert-to-project. List-only actions (new, reorder, filter) are not carried.
- Add `g` **go-to-project** on the view screen: when the task belongs to a project, replace the task view with that project's view (via `screen.Replace`, not a stacking push). Enabled only when the task has a project; no reciprocal task→ gesture (the project view already lists its tasks).
- Thread a `taskView` factory through `app.go` so the task list and task editor can navigate to the view without importing it directly, preserving the existing factory-injection pattern that avoids an import cycle.

## Capabilities

### New Capabilities
- `task-view-screen`: A read-display screen for a single task showing its fields and linked project, reached with enter from the task list, hosting the per-task action shortcuts and go-to-project navigation. Parallels `project-view-screen`.

### Modified Capabilities
- `task-list-query-ui`: enter opens the task view instead of the editor; a new `e` binding opens the editor; creating a task with `+`/`insert` navigates to the new task's view.
- `task-edit-ui`: on create (no ID), the editor dismisses and pushes the new task's view, via an injected view factory — mirroring `project-edit-ui`. Updates still dismiss only.

## Impact

- New package `tui/pages/tasks/taskview/` (`model.go`, `keymap.go`, `model_test.go`).
- `tui/pages/tasks/tasklist/{keymap,model,model_test}.go`: rebind enter/`e`, dispatch view vs edit, accept and use a `viewFn` factory for enter and new-task landing.
- `tui/pages/tasks/taskedit/model.go`: add a view-factory parameter; push the view on create. Ripples to every `taskedit.New` call site (list New passes the factory; edit/clarify pass nil).
- `tui/app.go`: hoist the existing `projectViewFn`, add a `taskViewFn`, thread both into the task list and editor.
- No domain, service, or storage changes; `TaskService.GetTask` already exists for the view's reload-on-init.
