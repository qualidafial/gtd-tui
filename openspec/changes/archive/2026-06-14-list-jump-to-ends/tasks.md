## 1. SQLite reorder generalization

- [x] 1.1 Refactor `sqlite/task.go` `shiftTask` to extract a `moveTaskTo(ctx, id, newPos int, filter)` helper that performs the neighbor-key lookup, `orderkey.Between`, and renumber fallback against an absolute target position; have `shiftTask` compute `pos` and delegate with `pos+delta`.
- [x] 1.2 Add `(*DB).MoveTaskFirst` and `(*DB).MoveTaskLast` in `sqlite/task.go`, wrapping `moveTaskTo` with `newPos = 0` and `newPos = len(filtered others)` inside `RunTx`; reject closed tasks and no-op at the boundary.
- [x] 1.3 Refactor `sqlite/project.go` `shiftProject` into a `moveProjectTo(ctx, id, newPos int, filter)` helper the same way, and add `(*DB).MoveProjectFirst` / `(*DB).MoveProjectLast`.
- [x] 1.4 Add SQLite tests: move-to-top/bottom within an empty filter, within a search filter (non-matching items keep order_keys), boundary no-op, closed-item rejection, and key-exhaustion renumber — for both tasks and projects.

## 2. Domain interfaces and service pass-throughs

- [x] 2.1 Add `MoveTaskFirst` / `MoveTaskLast` to the `TaskService` interface in `task.go` and `MoveProjectFirst` / `MoveProjectLast` to `ProjectService` in `project.go`, with doc comments mirroring the existing move methods.
- [x] 2.2 Implement the four pass-throughs in `service/task.go` and `service/project.go`.
- [x] 2.3 Add `MoveTaskFirst` / `MoveTaskLast` pass-throughs to the `ProjectTaskService` wrapper in `service/project_task.go`, preserving the project-scoping behavior of the existing wrapper move methods.
- [x] 2.4 Add a service/wrapper test confirming the project-task wrapper scopes move-to-top/bottom to the project (mirroring `TestProjectTaskService_MoveTaskUp_*`).

## 3. Task list keybindings

- [x] 3.1 Add `MoveFirst` (`shift+home`, help "move first") and `MoveLast` (`shift+end`, help "move last") bindings to `tui/pages/tasks/tasklist/keymap.go`, in the same help group as `MoveUp`/`MoveDown`.
- [x] 3.2 Extend `moveCmd` (or add a sibling) in `tasklist/model.go` to dispatch to `MoveTaskFirst` / `MoveTaskLast`, reusing the existing reload + reselect-by-id (`tasksReorderedMsg`) flow.
- [x] 3.3 Add Update cases for the new bindings and set their enablement in `updateKeybindings`: `MoveFirst` enabled exactly when `MoveUp` is, `MoveLast` exactly when `MoveDown` is.
- [x] 3.4 Add model tests: bindings enabled/disabled per selection (closed, first, last, interior) and that pressing them issues the correct move command and keeps the cursor on the moved task.

## 4. Project list keybindings

- [x] 4.1 Add `MoveFirst` / `MoveLast` (`shift+home` / `shift+end`) bindings to `tui/pages/projects/keymap.go`.
- [x] 4.2 Extend `projectlist.go` `moveCmd` to dispatch to `MoveProjectFirst` / `MoveProjectLast`, add Update cases, and set enablement in `updateKeybindings` aligned with the existing `MoveUp`/`MoveDown` predicates.
- [x] 4.3 Add project-list model tests mirroring the task-list ones (per-selection enablement and correct move dispatch).

## 5. Verification and docs

- [x] 5.1 Run `go build ./...` and `go test ./...`; confirm all packages pass.
- [x] 5.2 Update README keybinding tables (and any help-bar documentation) to list `shift+home` / `shift+end` for the task and project lists.
