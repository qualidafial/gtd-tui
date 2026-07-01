## 1. Date-editor overlay package

- [ ] 1.1 Create `tui/pages/tasks/taskdate/` package with a `Field` type (`Due`, `Defer`) that carries the field label ("Due" / "Defer Until") and knows which `Task` member it reads/writes.
- [ ] 1.2 Implement `Model` modeled on `taskstatus`: `New(task, svc, field)` builds `form.New(dateField, saveField)`, prefilling the date field from the task's current target value (empty when unset) and labeling it per `Field`.
- [ ] 1.3 Implement the commit command: copy the task, overwrite only the target member with the parsed `*time.Time` (nil when the field is empty), call `TaskService.UpdateTask`, then `screen.Dismiss()`; guard re-entry with an `applying` flag and emit an error on failure.
- [ ] 1.4 Implement `Update` (esc cancels, `form.SubmittedMsg` commits), `View` (header identifying the target date + form), `CapturingInput`, and `Keys` (form bindings + esc), mirroring `taskstatus`.
- [ ] 1.5 Add table-driven tests covering: prefill from an existing value, empty prefill when unset, set, change, clear-by-empty, that non-target attributes are preserved, and esc cancels without calling `UpdateTask`.

## 2. Task list bindings

- [ ] 2.1 Add `SetDue` (`d`, help "due") and `SetDefer` (`f`, help "defer") to `tasklist.KeyMap` in `keymap.go`, and include them in the appropriate `Keys()` group.
- [ ] 2.2 Add dispatch cases in `tasklist/model.go` that, for the selected task, `screen.Push(taskdate.New(task, svc, taskdate.Due/Defer))`.
- [ ] 2.3 Add/extend tasklist tests asserting `d` and `f` push the overlay with the correct `Field` for the selected task.

## 3. Task view bindings

- [ ] 3.1 Add `SetDue` (`d`) and `SetDefer` (`f`) bindings to `taskview.KeyMap` and its `Keys()` grouping.
- [ ] 3.2 Add dispatch cases in `taskview/model.go` pushing `taskdate.New(current, svc, field)` for each key.
- [ ] 3.3 Add/extend taskview tests asserting `d`/`f` push the overlay with the correct `Field`. Confirm (no code change needed) that the project view's embedded task list already routes `d`/`f` to the selected task.

## 4. Docs & verification

- [ ] 4.1 Update the README keybinding reference to list `d` (edit due) and `f` (edit defer) on the task list and task view.
- [ ] 4.2 Run `make` (build + full test suite) and manually verify set / change / clear for both Due and Defer from the task list and the task view.
