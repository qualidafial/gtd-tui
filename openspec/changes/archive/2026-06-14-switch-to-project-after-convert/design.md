## Context

The convert-to-project wizard (`tui/pages/tasks/taskconvert`) is launched from the task list via `screen.Push`, which the app handles by overlaying the wizard above the whole `tabcontainer` (`tui/app.go` `PushMsg` → `screen.Overlay`). On a successful commit it returns `screen.Dismiss()`, popping the overlay to reveal the `tabcontainer` on the Tasks tab. The just-converted task is gone from that list (it now belongs to the new project), so the user lands on a list that no longer reflects their work.

`ProjectService.ConvertTaskToProject(ctx, taskID, project, reframed) (gtd.Project, gtd.Task, error)` already returns the created project; today the wizard discards it and `convertedMsg` carries only `err`.

The project view (`tui/pages/projects/projectview`) is the established screen for a single project. It is constructed as `projectview.New(project, taskSvc, projectSvc, pickerFn)` and is pushed as a full-screen overlay above the `tabcontainer` from the project list (`projectlist.go`). The wizard does not have `taskSvc` or `pickerFn`, and the existing TUI pattern injects screen constructors as factory closures from `tui/app.go` (`pickerFn`, `convertFn`) rather than importing sibling page packages directly.

## Goals / Non-Goals

**Goals:**
- After a successful conversion, land the user on the new project's view.
- Reuse the existing factory-injection pattern; keep `taskconvert` decoupled from `projectview`/`tasklist`.
- Leave error and cancel paths untouched.

**Non-Goals:**
- No change to which tab is active underneath the overlay. The project view is an overlay above the `tabcontainer`; the underlying tab selection is irrelevant while it is shown and is out of scope. (Switching the underlying tab to Projects would require new tab-selection messaging and is not pursued.)
- No service, domain, schema, or query changes.

## Decisions

### Decision: Replace the wizard overlay with the project view
On success, return `screen.Replace(m.projectViewFn(project))` instead of `screen.Dismiss()`. `screen.Replace` swaps the active overlay in place (batching `RequestWindowSize` + `Init`), so the wizard overlay morphs directly into the project view overlay without briefly exposing the task list or adding a stack level. When the user later dismisses the project view, the overlay pops to the `tabcontainer` exactly as a dismissed wizard would have — so the back path is unchanged.

- *Alternative considered*: `screen.Dismiss(cmds...)` then push the project view. Rejected — pop-then-push briefly exposes the task list and adds a redundant stack frame; `Replace` exists precisely for this in-place swap.

### Decision: Inject a project-view factory, mirror `convertFn`/`pickerFn`
Add `projectViewFn func(gtd.Project) screen.Screen` to the wizard's `Model` and `New`. `tui/app.go` already holds `taskSvc`, `projectSvc`, and `pickerFn`; its `convertFn` closure constructs the wizard, so it builds the factory there:

```go
convertFn := func(task gtd.Task) screen.Screen {
    projectViewFn := func(p gtd.Project) screen.Screen {
        return projectview.New(p, taskSvc, projectSvc, pickerFn)
    }
    return taskconvert.New(task, projectSvc, projectViewFn)
}
```

This keeps `taskconvert` free of a `projectview`/`tasklist` import (consistent with how `pickerFn` and `convertFn` are threaded) and avoids any import-cycle risk.

- *Alternative considered*: import `projectview` directly inside `taskconvert` and pass `taskSvc`/`pickerFn` through. Rejected — diverges from the established injection pattern and couples the wizard to two more page packages.

### Decision: Carry the created project on `convertedMsg`
`convertCmd` captures the project returned by `ConvertTaskToProject` and puts it on `convertedMsg{project, err}`. `handleConverted` uses it only on the success branch (`err == nil`); the error branch is unchanged.

## Risks / Trade-offs

- **Underlying tab stays on Tasks.** While the project view is shown it is a full-screen overlay, so the tab bar is hidden and the mismatch is invisible. On dismiss the user returns to the Tasks tab — acceptable and matches today's dismiss target. Switching the tab is deliberately out of scope.
- **`nil` `projectViewFn` in tests.** Existing wizard tests call `New(task, projSvc)` / `New(task, nil)`. The signature gains a third parameter, so those call sites update to pass a factory (or `nil` where the success path is not exercised). The success-path test constructs a real project view factory to assert the landed screen.

## Migration

None — internal TUI wiring only.

## Open Questions

None.
