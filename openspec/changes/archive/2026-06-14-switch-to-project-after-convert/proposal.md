## Why

Converting a task to a project is a deliberate "this is actually bigger than one action" move — the user's attention is now on the new project. But the wizard currently dismisses back to the task list, where the source task has vanished (it now lives inside the project). The user is dropped on a list that no longer shows their work, with no confirmation of what was created and no path into it except switching tabs and hunting for the new project. The natural next step is to land on the new project so they can flesh it out.

## What Changes

- After a successful Convert to Project, the wizard SHALL open the newly created project's view in place instead of dismissing to the task list.
- The convert commit now surfaces the created `Project` to the wizard (the `ConvertTaskToProject` service call already returns it; only the TUI message needs to carry it) so the wizard can construct the project view.
- The convert-to-project wizard gains an injected project-view factory (mirroring the existing `pickerFn`/`convertFn` injection in `tui/app.go`), keeping the wizard decoupled from the `projectview`/`tasklist` packages.
- Error and cancel behavior are unchanged: a failed commit still shows the error in the wizard; Esc still leaves the task standalone and dismisses.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `task-project-restructure-ui`: the Convert to Project wizard's post-commit behavior changes from "dismiss" to "open the new project's view".

## Impact

- **TUI only.** `tui/pages/tasks/taskconvert/model.go`: add a `projectViewFn func(gtd.Project) screen.Screen` field and constructor parameter; carry the created `gtd.Project` on `convertedMsg`; on success `screen.Replace(projectViewFn(project))` instead of `screen.Dismiss()`.
- `tui/app.go`: the `convertFn` closure constructs the wizard with a project-view factory `func(p gtd.Project) screen.Screen { return projectview.New(p, taskSvc, projectSvc, pickerFn) }`.
- **No service, domain, schema, or query changes.** `ConvertTaskToProject` already returns `(Project, Task, error)`.
- **No new dependencies.**
