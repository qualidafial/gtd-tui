## 1. Tasklist API

- [x] 1.1 Add a `showProjectChip bool` parameter to `tasklist.New` in `tui/pages/tasks/tasklist/model.go`; store on `Model`
- [x] 1.2 Plumb the flag through to the delegate (extend `newDelegate` or store on `Model` and pass via render context)

## 2. Render

- [x] 2.1 In `tui/pages/tasks/tasklist/render.go`, extend `taskChips` (or add a parallel helper) to emit a `+<project title>` chip when the flag is set and the task's ProjectID resolves to a non-empty title
- [x] 2.2 Place the project chip after the assignee chip (new ordering: due → defer → assignee → project)
- [x] 2.3 Suppress the chip on dropped tasks (already handled by the early return); keep it on done tasks
- [x] 2.4 Pick a distinct color for the project chip and add it to `chipColors` / `newChipColors`

## 3. Call sites

- [x] 3.1 In `tui/app.go`, pass `true` for the global tasks tab's tasklist
- [x] 3.2 In `tui/pages/projects/projectview/model.go`, pass `false` for the in-project tasklist

## 4. Tests

- [x] 4.1 Add a `tui/pages/tasks/tasklist/render_test.go` case: open task with ProjectID renders `+<title>` chip
- [x] 4.2 Add: standalone task renders no project chip
- [x] 4.3 Add: in-project tasklist (flag off) renders no project chip even when ProjectID is non-nil
- [x] 4.4 Add: dropped task with ProjectID shows no chips; done task with ProjectID still shows the project chip
- [x] 4.5 Add: project chip is placed after the assignee chip in the rendered order

## 5. Verification

- [x] 5.1 `go build ./...` clean
- [x] 5.2 `go test ./...` green
- [x] 5.3 Manual TUI check: global tasks tab shows project chips on tasks that have a project; project view's task list does not