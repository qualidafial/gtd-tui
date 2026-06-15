## 1. taskedit view-factory support

- [x] 1.1 Add a `ViewFactory func(gtd.Task) screen.Screen` type to `taskedit` and a parameter on `taskedit.New` (after `projectName`); store it on the model. Mirror `projectedit`'s factory shape.
- [x] 1.2 On successful save: if the task had no ID and a view factory is set, `screen.Replace` with the new task's view; otherwise dismiss only. Mirror `projectedit.handleSaved`.
- [x] 1.3 Update `taskedit/model_test.go`: assert create-with-factory replaces with the view, create-without-factory and update dismiss.

## 2. taskview screen

- [x] 2.1 Create `tui/pages/tasks/taskview/keymap.go`: `KeyMap` with `Edit`, `ToggleComplete`, `Drop`, `AssignToProject`, `ConvertToProject`, `GoToProject` bindings (`e`, `space`, `delete`, `p`, `c`, `g`) and `Keys()` returning them as a help group. Mirror `projectview/keymap.go`.
- [x] 2.2 Create `tui/pages/tasks/taskview/model.go`: `Model` holding the task, `taskSvc`, `projectNameFn`, `pickerFn`, `convertFn`, `projectViewFn`, and `KeyMap`. `New(...)` constructs it and calls `updateKeybindings`.
- [x] 2.3 Implement `Init` → `reloadCmd` (re-`GetTask` by ID) and a `taskReloadedMsg` handler that stores the task and reconciles keybindings. Reuse the `projectview.reloadCmd` shape.
- [x] 2.4 Implement `renderHeader`/`View`: title, status, project (`+<name>` via `projectNameFn`, omitted when standalone), assignee (when delegated), due, description; suppress empty optionals. Reuse `projectview`'s label/value styles.
- [x] 2.5 Implement `Update` key handling: `e` push `taskedit.New(task, svc, projectName, nil)`; `space` push `taskstatus` (complete/reopen by status); `delete` push `taskstatus` drop; `p` push `pickerFn(task)`; `c` push `convertFn(task)`; `g` `screen.Replace(projectViewFn(project))`.
- [x] 2.6 Implement `updateKeybindings`: enable `Drop` only when open, `ConvertToProject` only when standalone, `GoToProject` only when the task has a project, `ToggleComplete` help label complete/reopen by status.
- [x] 2.7 Esc-dismiss is handled by the overlay fallback (no explicit handler, matching `projectview`); `Keys` returns the keymap group. No `CapturingInput` needed — the view captures no input.
- [x] 2.8 Add `taskview/model_test.go`: header rendering (all fields / standalone-omits / status labels), keybinding guards, toggle label, and the `e`/`space`/`delete`/`p`/`c`/`g` dispatches plus reload.

## 3. Task list rebind

- [x] 3.1 In `tasklist/keymap.go`: change `Edit` binding to `e`, add a `View` binding on `enter` with help "view", and include both in `Keys()`.
- [x] 3.2 In `tasklist/model.go`: add a `viewFn ViewFactory` field and `New` parameter; dispatch `View` (enter) → `screen.Push(m.viewFn(ti.task))` and `Edit` (`e`) → push editor; pass the view factory into the new-task `taskedit.New` and `nil`/`""` view-factory for the edit call. Enter falls back to the editor when no view factory (in-project list).
- [x] 3.3 Update `tasklist/model_test.go`: enter pushes the view; enter falls back to the editor without a factory; `e` pushes the editor; new-task path threads the view factory.

## 4. app.go wiring

- [x] 4.1 Hoist `projectViewFn` out of the `convertFn` closure into a shared local in `app.go`.
- [x] 4.2 Add `taskViewFn := func(t gtd.Task) screen.Screen { return taskview.New(t, taskSvc, projectNameFn, pickerFn, convertFn, projectViewFn) }`.
- [x] 4.3 Pass `taskViewFn` into `tasklist.New` for the Tasks tab.

## 5. Verify

- [x] 5.1 `go build ./...` and `go test ./...` pass.
- [x] 5.2 Behavior verified by screen-level tests driving the real `Update` logic: enter→view / `e`→edit (tasklist), create→replace-with-view (taskedit), and every taskview dispatch + guard (`e/space/delete/p/c/g`, drop/convert/go-to enablement, reload). `cmd/gtd` binary links. Interactive TTY run not available in this environment.
- [x] 5.3 README task-list table updated (enter→view, add `e`) and a Task view keybinding section added.