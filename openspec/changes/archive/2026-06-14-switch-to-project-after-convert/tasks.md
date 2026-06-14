## 1. Wizard carries the created project and lands on its view

- [x] 1.1 Add `projectViewFn func(gtd.Project) screen.Screen` to `taskconvert.Model` and to `New`, storing it on the model.
- [x] 1.2 Add a `project gtd.Project` field to `convertedMsg` and set it from the `ConvertTaskToProject` return value in `convertCmd`.
- [x] 1.3 In `handleConverted`, on the success branch return `screen.Replace(m.projectViewFn(msg.project))` instead of `screen.Dismiss()`; leave the error branch unchanged.

## 2. App wiring

- [x] 2.1 In `tui/app.go`, update the `convertFn` closure to build a project-view factory `func(p gtd.Project) screen.Screen { return projectview.New(p, taskSvc, projectSvc, pickerFn) }` and pass it as the third arg to `taskconvert.New`; add the `projectview` import.

## 3. Tests

- [x] 3.1 Update existing `taskconvert` test call sites for the new `New` signature (pass `nil` or a stub factory where the success path is not exercised).
- [x] 3.2 Add/adjust a test asserting that on a successful commit the wizard's resulting screen is the project view of the created project (factory invoked with the returned project), not a dismiss.
- [x] 3.3 Confirm the commit-error test still asserts the wizard stays open and the project view is not opened.

## 4. Verification

- [x] 4.1 Run `go build ./...` and `go test ./tui/...` (and the service tests touched, if any) — all green.
- [x] 4.2 Manually verify in the running app: convert a standalone task and confirm you land on the new project's view; cancel still returns to the task list; a forced commit error still shows in the wizard.
