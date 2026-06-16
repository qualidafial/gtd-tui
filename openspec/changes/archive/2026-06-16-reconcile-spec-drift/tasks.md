## 1. task-list-presentation

- [x] 1.1 Confirm `tui/pages/tasks/tasklist/render.go` renders the project as a leading label (no `+` prefix) ahead of the title, suppressed on dropped tasks and kept on done — matches Requirement: Project label
- [x] 1.2 Confirm `theme.Project` is indigo (`#7571F9`, `theme.LogoBg`) and that the label brightens with selection — matches the color and Selection highlight scope requirements
- [x] 1.3 Confirm `taskChips` no longer emits a project chip and chip order is due/defer/assignee — matches Chip ordering and alignment + Urgency colors
- [x] 1.4 Confirm the existing render tests (`render_test.go`: `TestTaskChips`, `TestProjectLabel`) cover the reconciled scenarios

## 2. project-view-screen

- [x] 2.1 Confirm `app.go` wires a task `ViewFactory` into the project view's embedded tasklist so `enter` pushes the task view and `e` still edits — matches Open task view from project view
- [x] 2.2 Confirm the in-project task view is constructed with a nil `ProjectViewFactory` so go-to-project (`g`) is disabled

## 3. form-field-toolkit

- [x] 3.1 Confirm `selectfield.Model.SetOptions`, `WithHideWhenEmpty`, and `form.Model.UpdateField` exist with the signatures described in the spec
- [x] 3.2 Confirm `selectfield_test.go` / `form_test.go` cover SetOptions, hide-when-empty visibility, and in-place UpdateField

## 4. domain-model

- [x] 4.1 Confirm the `gtd.Task` struct has no `Kind` field and no `TaskKind` type exists — delegation is inferred from a non-nil `Assignee`

## 5. Finalize

- [x] 5.1 Run `openspec validate reconcile-spec-drift --strict`
- [x] 5.2 Run `go test ./...` and `go build ./...` to confirm no code drift was introduced
- [x] 5.3 Archive the change with `/opsx:archive` once verified
