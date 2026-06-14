## Context

Task-vs-project classification happens once, in the clarify wizard (`InboxService.ClarifyAsTask` / `ClarifyAsProject`), and the originating `Item` is consumed. After clarify there is no way to change a task into a project or vice versa. These three actions are post-clarify re-classifications operating purely on materialized `Task`/`Project` entities — no `Item` is involved, so there is no `ClarifiedInto` bookkeeping.

Current state:
- `Task.ProjectID *int64` — nullable; standalone when nil. At most one project per task.
- `Project` carries `Title`, `Outcome`, `Description`, `Due`, `Status` (open/someday/done/dropped), plus ordering rank.
- `Task` carries `Title`, `Description`, `Status` (open/done/dropped), `Assignee`, `Due`, `DeferUntil`, plus filter-scoped ordering rank. Task has **no** `Outcome` field and no `someday` status.
- All services hold a `*sqlite.DB` exposing every store plus `RunTx`. `InboxService` is the precedent for cross-store transactional orchestration.
- An existing `project-picker-overlay` already assigns a project to a task (task→project direction). Link Task is the reverse-direction UX (project→task) and needs its own overlay.

## Goals / Non-Goals

**Goals:**
- Let the user promote a standalone task into a project, collapse an empty project into a task, and link a standalone task into a project — all atomically and losslessly.
- Reuse existing patterns: the form toolkit for the convert wizard, the picker-overlay pattern for task selection, the service `RunTx` pattern for atomicity.
- Keep clarify and query behavior untouched.

**Non-Goals:**
- No new orchestrator service (scope was deliberately collapsed; methods live on existing services).
- No schema migration — every operation is expressible with existing columns.
- No re-parenting between projects (Link Task accepts standalone tasks only).
- No conversion of someday/done/dropped projects to tasks (open-only).
- No batch/multi-select restructuring; one entity at a time.

## Decisions

### Decision: Three transactional service methods, no new orchestrator
Add `ConvertTaskToProject`, `ConvertProjectToTask`, `LinkTaskToProject` backed by `*sqlite.DB` + `RunTx`. The proposal collapsed #1/#2 into one action, so a dedicated `RestructureService` would be ceremony. Placement: `ConvertTaskToProject` and `ConvertProjectToTask` belong on `ProjectService` (project-creating / project-destroying); `LinkTaskToProject` on `ProjectService` too (it is invoked from the project view and conceptually "the project adopts a task"). All three guard their invariants inside the transaction so concurrent races still terminate safely.

- *Alternative considered*: reuse `UpdateTask` for Link. Rejected — the standalone-only and project-exists guards plus the bottom-of-list rank placement want a named, testable method.

### Decision: Convert to Project is form-first, commit-at-end
The clarify checkpoint pattern exists because the source `Item` is fragile. Here the source task already exists and is safe: abandoning the wizard leaves the task standalone, losing nothing. So the wizard collects all fields first (project Title/Outcome/Description, re-scoped task Title/Description) and commits in a single `ConvertTaskToProject` transaction on submit. `ConvertTaskToProject(ctx, taskID, project, reframedTask)` in one tx: create the project (`Status=open`), set `reframedTask.ProjectID = newProject.ID`, apply the reframed fields, persist the task. Returns `(Project, Task, error)`.

- Field flow: `task.Title`/`task.Description` pre-populate **both** the project fields and the (editable) reframed-task fields; `project.Outcome` starts empty and is required; the user narrows the task title down to the next action.
- *Alternative considered*: early-checkpoint like clarify. Rejected — unnecessary complexity when the source is durable.

### Decision: Convert to Task guards open + zero tasks, folds Outcome into Description
`ConvertProjectToTask(ctx, projectID) (Task, error)` in one tx: verify the project is `open` and has **zero tasks of any status** (the only lossless case — the relationship spec keeps done/dropped tasks attached as history, so a non-empty project cannot collapse without loss); create a standalone task (`ProjectID=nil`, `Status=open`) inheriting `Title`, `Description`, `Due`; if `Outcome` is non-empty, append it to the task `Description` so nothing is lost; delete the project. No new fields from the user — a plain confirm in the UI suffices.

- *Alternative considered*: allow any status / zero-pending-tasks. Rejected per proposal — loses history and multiplies edge cases (someday has no task equivalent; done/dropped empty projects are clutter the user can delete directly).

### Decision: Link Task via a new selection-only `taskpicker` overlay
`LinkTaskToProject(ctx, taskID, projectID) (Task, error)` in one tx: verify the task is standalone (`ProjectID == nil`) and the project exists; set `task.ProjectID = projectID`; place the task at the **bottom** of the project's task ordering. The TUI gets a new `tui/pages/projects/taskpicker/` overlay: a `selectfield` of standalone open tasks that is **selection-only** — on confirm it emits a selection message and dismisses, and the **project view** applies it by calling `LinkTaskToProject` and owns any error. The picker calls no mutating operation itself.

- Broadcast mechanism: `screen.Dismiss(cmds.Emit(selectedMsg{task}))`. `Dismiss` sequences the emit *after* the pop, so the selection message reaches the project view once it is the active screen (it cannot be absorbed by the dismissed sentinel).
- Diverges deliberately from the existing self-contained `project-picker-overlay`, which still calls `UpdateTask` itself. A separate change will later convert both pickers to this select-and-broadcast shape; this change establishes the pattern on the new picker without touching shipped code.
- Action gating: the picker is opened only when at least one standalone open task exists. The project view disables the Link Task action when there are no candidates, so the picker never renders an empty state. This requires the project view to know whether any standalone open task exists (a count/existence query against TaskService).
- Someday side effect: linking into a someday project removes the task from default views (`IncludeSomedayProjects` rule). Allowed without a confirm — it is correct, reversible, and the project view still shows the task — but the spec documents it so it is not mistaken for a bug.
- *Alternative considered*: reuse the existing project-picker (task→project). Rejected — the user wants the action originating from the project view, which needs the reverse picker.

### Decision: Entry points
- **Convert to Project**: task list only (acts on the selected standalone task; non-standalone tasks do not expose the action).
- **Convert to Task**: project view and project list (acts on the open, empty project; the action is hidden/disabled when guards fail).
- **Link Task**: project view only.

## Risks / Trade-offs

- **"Where did my task go?" on Link into someday** → documented in the spec; the project view shows the task, and the action is reversible by reopening the project or unlinking.
- **Convert to Task is destructive (deletes a project row)** → guarded to empty projects only, so no task/history is lost; the resulting task is a clean inverse of Convert to Project.
- **Outcome folded into Description is one-directional** → converting task→project→task will not perfectly round-trip the Outcome (it lands in Description). Acceptable: an empty project's outcome is typically throwaway, and folding is lossless even if not structure-preserving.
- **Action visibility logic duplicated across three pages** → keep the guard predicates (`task.ProjectID == nil`, `project open && task count == 0`) in one place (domain/service helpers) and have the pages call them, rather than re-deriving.

## Open Questions

- Should Convert to Task prompt a confirm dialog, or convert immediately? Leaning confirm (it deletes a project), but a single keystroke with an undo-via-Convert-to-Project may be acceptable. Resolve during apply.
- Exact home of the convert-to-project wizard package (`tui/pages/tasks/taskconvert/` vs. folding into an existing task page). Resolve during apply.
