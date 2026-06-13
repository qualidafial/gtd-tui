## 1. Thread ProjectService into the wizard

- [x] 1.1 `tui/app.go`: pass the existing `projectSvc` to `inbox.New`
- [x] 1.2 `tui/pages/inbox/model.go`: add a `projectSvc gtd.ProjectService` field, accept it in `inbox.New`, and pass it to `clarify.New`
- [x] 1.3 `tui/pages/inbox/clarify/model.go`: add a `projectSvc gtd.ProjectService` field and accept it in `clarify.New`

## 2. Loading phase for open projects

- [x] 2.1 Add a `ready bool` flag and a `loadProjectsCmd` issued from `Init` that calls `ProjectService.ListProjects(open)`
- [x] 2.2 Add a `projectsLoadedMsg` handler that builds the initial form with the loaded options and sets `ready = true`
- [x] 2.3 `View` renders `Loading…` until `ready`; `CapturingInput`/`Keys` return their not-ready forms until then (mirror the `projectpicker` pattern)
- [x] 2.4 On load error, surface it via the existing error path

## 3. Project select in the single-task block

- [x] 3.1 `buildInitialForm` takes the loaded `[]gtd.Project`; add a `selectfield` "Project" with `WithNone("(none)")` listing open projects, defaulting to `(none)`
- [x] 3.2 Make the field visible only when `actionable && !multiStep` (single-task branch), independent of the `<2 min`/doer answers
- [x] 3.3 Place the field last in the single-task block, after the assignee field, per the per-task block ordering

## 4. Wire the selection through to ClarifyAsTask

- [x] 4.1 `taskFromVals` reads the `"project"` value; when present and non-zero, set `task.ProjectID`; default to standalone when the key is absent (loop form) and comment the default
- [x] 4.2 Confirm the single-task path still routes to `ClarifyAsTask` unchanged and does NOT enter `phaseProjectLoop`
- [x] 4.3 Confirm the sub-2-minute do-it-now path carries the chosen `ProjectID` into the committed (open, then completed) task

## 5. Tests

- [x] 5.1 Single-task clarify with a selected project: the created task has `ProjectID` set to the chosen project
- [x] 5.2 Single-task clarify left on `(none)`: the created task has nil `ProjectID`
- [x] 5.3 Sub-2-minute do-it-now with a selected project: task is created open with the `ProjectID`, then completed
- [x] 5.4 Empty project list: only `(none)` is offered and the task is created standalone
- [x] 5.5 Loop form path is unaffected (loop tasks attach to the new project, not to a stray select value)

## 6. Spec sync

- [x] 6.1 Update `inbox-page` per the change's delta (single-task select default `(none)`; wizard loads open projects)
- [x] 6.2 Update `project-picker-overlay` to describe the `selectfield`-based picker instead of `huh.Select`
- [x] 6.3 `openspec validate clarify-task-existing-project --strict` passes
