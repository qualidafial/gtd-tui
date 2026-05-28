## 1. Project edit overlay

- [x] 1.1 Create `tui/pages/projects/projectedit/model.go` with huh form (Title, Outcome, Description, Due), read-only header for existing projects, create-vs-update branching on ID==0, save-error surfacing with esc-to-retry
- [x] 1.2 On create success, use `tea.Sequence` to dismiss then push project view via `ViewFactory` func; on update success, dismiss only
- [x] 1.3 Add tests for projectedit: create save, update save, validation (empty title, empty outcome), save error retry, esc cancel

## 2. Wire into project list

- [x] 2.1 Add `e` key binding to `keymap.go`, enabled when a project is selected
- [x] 2.2 In `projectlist.go`, handle `e` key to push `projectedit` with the selected project
- [x] 2.3 Change `+`/`insert` handler to push `projectedit` with an empty project instead of `projectcreate`
- [x] 2.4 Pass a `ViewFactory` closure to `projectedit` that constructs a `projectview.Model`

## 3. Wire into project view

- [x] 3.1 Add `e` key binding to project view, enabled when task list is not capturing input
- [x] 3.2 Handle `e` key to push `projectedit` with the current project (no ViewFactory needed—update only)
- [x] 3.3 Reload project from service on overlay dismiss to refresh the header

## 4. Cleanup

- [x] 4.1 Delete `tui/pages/projects/projectcreate/` package
- [x] 4.2 Remove `projectcreate` import from `projectlist.go`