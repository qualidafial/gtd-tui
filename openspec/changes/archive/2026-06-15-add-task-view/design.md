## Context

Projects already have a view screen (`tui/pages/projects/projectview`) reached with enter from the project list, with `e` opening the editor. The task list has no such split: enter opens the editor directly and there is no read-display screen. This change brings tasks to parity and creates the home where planned task features (comments, history) will live.

The structural parallel to `projectview` is shallow. `projectview`'s body is an embedded `tasklist`, and most of its code services that child (window-size forwarding, `statsCmd`, link/convert guards). A task has no embedded collection yet, so `taskview` is essentially "render the header fields, handle a few per-task action keys, reload the task on init." It should copy `projectview`'s *conventions* (a `reloadCmd` on `Init`, `updateKeybindings`, `Keys()`), not its structure.

Key existing mechanism: the app calls `Init()` on the revealed screen after every `DismissMsg` (`tui/app.go:96-99`). So any overlay that dismisses back to the view triggers `taskview.Init`, which re-runs `reloadCmd` and refreshes the displayed task for free — no explicit per-action reload plumbing is needed.

## Goals / Non-Goals

**Goals:**
- A read-display task view screen showing the task's fields and linked project name.
- Task list parity with the project list: enter → view, `e` → edit, create → land on view.
- Carry the per-task actions that make sense (`e`, `space`, `delete`, `p`, `c`) into the view.
- `g` go-to-project that replaces (not stacks) the task view with the project's view.
- Preserve the factory-injection pattern so no import cycle is introduced.

**Non-Goals:**
- Tab chrome / multiple tabs (added later when a second tab exists).
- Comments, history, or any new entity — this is just the host screen.
- Editing fields inline on the view (editing stays in `taskedit`).
- A reciprocal project→task navigation gesture.
- Any domain/service/storage change.

## Decisions

### Decision: `taskview` reaches sibling overlays directly but cross-package screens via injected factories

The naive wiring is a cycle:

```
tasklist ──enter pushes──▶ taskview ──g go-to──▶ projectview ──embeds──▶ tasklist
```

The codebase already breaks this with factory injection in `app.go`: `pickerFn`, `convertFn`, and `projectViewFn` are `func(...) screen.Screen` built once and threaded down, so `tasklist` imports neither `projectpicker` nor `projectview`. `taskview` follows the same rule:

- Imports its own sibling overlays directly: `taskedit`, `taskstatus` (under `tui/pages/tasks/`, no cycle).
- Reaches `projectview` (for `g`) only through an injected `projectViewFn` — which already exists in `app.go` (taskconvert uses it).
- Receives `pickerFn` (assign) and `convertFn` (convert) as the task list does.

`tasklist` likewise gains a `viewFn ViewFactory` rather than importing `taskview`. Result: `tasklist → taskview → projectview → tasklist` has no compile-time edge, because each arrow that would close the loop is a runtime func value built in `app.go`.

**Alternative considered:** have `taskview` import `projectview` directly and break the cycle by moving the embedded list out of `projectview`. Rejected — far more invasive and abandons the established pattern.

### Decision: reload via `Init`-on-dismiss, not per-action reload commands

`taskview.Init` batches a `reloadCmd` (re-`GetTask` by ID, like `projectview.reloadCmd`). Because the app re-inits the revealed screen on `DismissMsg`, every overlay that finishes by dismissing (`taskedit`, `taskstatus`, `projectpicker`) refreshes the view automatically. No `linkedMsg`/`convertedMsg`-style plumbing is needed in `taskview` for those.

**Exception — convert to project:** `taskconvert` finishes with `screen.Replace(projectViewFn(...))` (`taskconvert/model.go:121`), so from the view it swaps the task view for the new project's view. Correct: the task is gone, the user lands on what it became.

### Decision: `g` go-to-project uses `screen.Replace`

Per the user: replace rather than stack, to avoid deep stacks. `g` is enabled only when `task.ProjectID != nil`. From the standalone list path it swaps task-view→project-view over the list; from inside a project's embedded list it swaps to that same project's view (mildly redundant but harmless). No reciprocal gesture — the project view already lists its tasks.

### Decision: new-task landing mirrors `project-edit-ui`

`taskedit.New` gains a view-factory parameter. On create (no ID) with a factory it replaces itself with the new task's view via `screen.Replace`, exactly as `projectedit` does (the `Replace` helper morphs to the next screen and batches its window-size + `Init`). Updates, and creates without a factory, dismiss only. The task list's `New` passes `taskViewFn`; the list's `Edit` and any clarify call sites pass `nil`. `taskview`'s own `e` also passes `nil` (editing an existing task never creates).

### Decision: header field set

Render, in order, suppressing empty optionals: Title (bold), Status, Project (`+<name>` via `projectNameFn`, omitted when standalone), Assignee (when delegated), Due, Description. Reuse `projectview`'s label/value style approach.

## Risks / Trade-offs

- **Muscle-memory break: enter no longer edits** → Acceptable and intended; it matches the project list, and `e` edits from both list and view.
- **`taskedit.New` signature change ripples to all call sites** → Mechanical; mirrors the existing `projectedit` change. Enumerate call sites in tasks (list New/Edit, inbox clarify flows, taskview).
- **`g` redundancy when already inside a project view** → Harmless swap to the same project; not worth special-casing.
- **`taskview` constructor has many params** (task, taskSvc, projectNameFn, pickerFn, convertFn, projectViewFn) → Consistent with `projectview`'s constructor; keep it as a flat factory in `app.go`.

## Migration Plan

Pure additive UI wiring; no data migration. Rollback is reverting the commit. `TaskService.GetTask` already exists (`service/task.go:19`), so the view's reload needs no new service method.

## Open Questions

None — the three design questions (new-task landing, no tab chrome, `p`/`g` semantics) are resolved.