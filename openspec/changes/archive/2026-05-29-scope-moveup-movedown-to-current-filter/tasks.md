## 1. Interface signatures

- [x] 1.1 Add `filter TaskFilter` parameter to `MoveTaskUp` and `MoveTaskDown` on `gtd.TaskService` in `task.go`
- [x] 1.2 Add `filter ProjectFilter` parameter to `MoveProjectUp` and `MoveProjectDown` on `gtd.ProjectService` in `project.go`

## 2. SQLite implementation — tasks

- [x] 2.1 Extract the `TaskFilter` → WHERE-clause translation from `sqlite.DB.ListTasks` into a private helper (e.g. `applyTaskListFilter(q sq.SelectBuilder, f gtd.TaskFilter) sq.SelectBuilder`); refactor `ListTasks` to use it (no behavior change)
- [x] 2.2 Extend `pendingOrder` to take a `gtd.TaskFilter` (status forced to open); single helper serves both the filtered-neighbor query and the all-open renumber query
- [x] 2.3 Thread `filter` through `MoveTaskUp`/`MoveTaskDown` → `shiftTask(ctx, id, delta, filter)`; compute `pos`, `newPos`, `prevKey`, `nextKey` from the filtered set; keep the `orderkey.Between` fast path; on exhaustion call `pendingOrder` with an empty filter to load all open tasks, splice the moving id after its filtered prev neighbor, assign evenly-spaced keys to every row

## 3. SQLite implementation — projects

- [x] 3.1 Extract the `ProjectFilter` → WHERE-clause translation from `sqlite.DB.ListProjects` into `applyProjectListFilter`; refactor `ListProjects` to use it
- [x] 3.2 Extend `projectOrderByStatus(ctx, status, excludeID)` to take a `gtd.ProjectFilter` (status always forced to the supplied status); single helper serves both filtered-neighbor and all-same-status queries
- [x] 3.3 Thread `filter` through `MoveProjectUp`/`MoveProjectDown` → `shiftProject(ctx, id, delta, filter)` with the same fast-path/renumber logic as tasks (renumber spans the entire same-status group on exhaustion)

## 4. Service layer

- [x] 4.1 Update `service.TaskService.MoveTaskUp`/`MoveTaskDown` to accept and forward `filter`
- [x] 4.2 Update `service.ProjectService.MoveProjectUp`/`MoveProjectDown` to accept and forward `filter`
- [x] 4.3 Update `service.ProjectTaskService.MoveTaskUp`/`MoveTaskDown` to set `filter.ProjectID = &s.projectID` before delegating (mirror `ListTasks`)

## 5. TUI call sites

- [x] 5.1 In `tui/pages/tasks/tasklist/model.go::moveCmd`, pass `m.filter` into `svc.MoveTaskUp`/`svc.MoveTaskDown`
- [x] 5.2 In `tui/pages/projects/projectlist.go::moveCmd`, pass `m.filter` into `svc.MoveProjectUp`/`svc.MoveProjectDown`

## 6. Tests

- [x] 6.1 Update `sqlite/task_test.go` move tests to pass an empty filter, verifying the global-pending behavior still holds for that case
- [x] 6.2 Add `sqlite/task_test.go` cases for: move down within a search filter (filtered neighbor used), move on a one-item filter is a no-op, in-project filter swaps only within that project
- [x] 6.3 Update `sqlite/project_test.go` move tests to pass an empty `ProjectFilter`; add a search-filtered case
- [x] 6.4 Add `service/project_task_test.go` cases for the wrapper's ProjectID injection on MoveTask (empty filter and foreign-ProjectID override); TaskService/ProjectService Move methods are thin pass-throughs covered by DB tests
- [x] 6.5 TUI model tests reference local `keys.MoveUp/MoveDown` bindings (unchanged); no test updates needed for the signature rename

## 7. Verification

- [x] 7.1 `go build ./...` clean
- [x] 7.2 `go test ./...` green
- [x] 7.3 Manual TUI check: with a filter narrowing the projects tab, one shift+down visibly moves the project by one slot; same for the in-project task list with a filter
