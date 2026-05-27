## 1. Create key migration

- [x] 1.1 Change task list create keybinding from `n` to `+`/`insert` in tasklist/keymap.go and tasklist/model.go
- [x] 1.2 Change project list create keybinding from `n` to `+`/`insert` in projectlist.go and projectcreate references
- [x] 1.3 Update keybinding state logic: disabled-when-empty checks reference `+` instead of `n`

## 2. Project task service wrapper

- [x] 2.1 Implement `projectTaskService` struct in service layer wrapping `gtd.TaskService` with a project ID
- [x] 2.2 `ListTasks` injects ProjectID into the filter before delegating
- [x] 2.3 `CreateTask` stamps ProjectID on the task before delegating
- [x] 2.4 All other methods delegate unchanged
- [x] 2.5 Add tests for wrapper (list scoping, create stamping, passthrough)

## 3. Project picker overlay

- [x] 3.1 Implement `projectpicker.Model` in `tui/pages/projects/projectpicker/` with huh.Select over open projects + "(none)"
- [x] 3.2 Pre-select current project (or "(none)" if nil ProjectID)
- [x] 3.3 On confirm, call `TaskService.UpdateTask` to set/clear ProjectID; skip update if unchanged
- [x] 3.4 Display error on save failure; esc dismisses without change
- [x] 3.5 Add tests for picker (assign, unlink, no-op, cancel, error)

## 4. Wire picker into task list

- [x] 4.1 Add picker factory `func(gtd.Task) screen.Screen` parameter to `tasklist.New`
- [x] 4.2 Add `p` keybinding to tasklist keymap; disabled when no task selected or no factory provided
- [x] 4.3 Handle `p` key in tasklist Update: push picker overlay via factory
- [x] 4.4 Wire factory from app.go, constructing `projectpicker.New` with ProjectService and TaskService
- [x] 4.5 Add test for `p` key pushing picker overlay

## 5. Project view screen

- [x] 5.1 Implement `projectview.Model` in `tui/pages/projects/projectview/` with project header and embedded tasklist
- [x] 5.2 Render header: title, status, outcome, due — only non-empty fields
- [x] 5.3 Calculate header height and pass remaining space to embedded tasklist via WindowSizeMsg
- [x] 5.4 Construct embedded tasklist with projectTaskService wrapper and empty default query
- [x] 5.5 Pass picker factory to embedded tasklist so `p` works for reassignment
- [x] 5.6 Forward Init/Update/View/KeyMap to embedded tasklist (below header)
- [x] 5.7 Add tests for project view (header rendering, task scoping, create inherits project)

## 6. Wire project view into project list

- [x] 6.1 Add `enter` keybinding to project list keymap
- [x] 6.2 Handle `enter` key in projectlist Update: push projectview overlay for selected project
- [x] 6.3 Pass TaskService and ProjectService to project list so it can construct project view
- [x] 6.4 Update app.go to pass both services to project list
- [x] 6.5 Add test for enter key pushing project view