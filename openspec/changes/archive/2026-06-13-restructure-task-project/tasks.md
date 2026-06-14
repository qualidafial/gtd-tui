## 1. Service layer — restructuring operations

- [x] 1.1 Add `ConvertTaskToProject(ctx, taskID int64, project gtd.Project, reframedTask gtd.Task) (gtd.Project, gtd.Task, error)` to `ProjectService`, implemented over `RunTx`: reject non-standalone source task, default project Title/Description from the task, create project (`Status=open`), set reframed task's ProjectID to the new project, persist the reframed task.
- [x] 1.2 Add `ConvertProjectToTask(ctx, projectID int64) (gtd.Task, error)` to `ProjectService`, over `RunTx`: reject non-open status, reject any tasks of any status, create standalone task (`ProjectID=nil`, `Status=open`) inheriting Title/Description/Due, fold non-empty Outcome into Description, delete the project.
- [x] 1.3 Add `LinkTaskToProject(ctx, taskID, projectID int64) (gtd.Task, error)` to `ProjectService`, over `RunTx`: reject non-standalone task, reject invalid project, set ProjectID, place at bottom of the project's task ordering.
- [x] 1.4 Extend the `gtd.ProjectService` interface in `project.go` with the three new methods and doc comments mirroring the existing style.
- [x] 1.5 Add any required sqlite helper(s) for ordering a newly linked task at the bottom of a project's tasks (reuse the existing rank/order machinery used by MoveTask/CreateTask).

## 2. Service layer — tests

- [x] 2.1 Test `ConvertTaskToProject`: happy path (project created, task re-parented, fields default from task), rejects task already in a project, field overrides honored, atomicity (failure leaves no project and unchanged task). Use the real service + `:memory:` DB.
- [x] 2.2 Test `ConvertProjectToTask`: happy path (task created standalone, project deleted, Due carried over), Outcome folded into Description, rejects non-open status, rejects project with done/dropped/open tasks, atomicity.
- [x] 2.3 Test `LinkTaskToProject`: happy path (ProjectID set, ordered last), rejects non-standalone task, rejects invalid project, someday-project task excluded from default views and restored on ReopenProject.

## 3. Task picker overlay (selection-only)

- [x] 3.1 Create `tui/pages/projects/taskpicker/`: a `selectfield` of standalone open tasks loaded from TaskService, sized to the overlay, no `(none)` option. No empty-state — the caller only opens it when candidates exist.
- [x] 3.2 On confirm, emit a `selectedMsg{task}` via `screen.Dismiss(cmds.Emit(...))` (emit sequenced after the pop) and dismiss; call no mutating service. esc dismisses without emitting. The picker owns no apply error.
- [x] 3.3 Add overlay tests covering candidate filtering (standalone open only), confirm-emits-selection-and-dismisses, esc-cancels-without-emit, and that no mutating service is invoked.

## 4. Convert-to-Project wizard

- [x] 4.1 Create the convert-to-project wizard (form-toolkit, form-first) collecting project Title/Outcome/Description and reframed task Title/Description; pre-populate via `SetValue` from the source task (Outcome empty); decide final package home (e.g. `tui/pages/tasks/taskconvert/`).
- [x] 4.2 On submit, call `ProjectService.ConvertTaskToProject`; dismiss on success; display errors in-wizard; esc/abandon leaves the task unchanged (no checkpoint).
- [x] 4.3 Add wizard tests: pre-population, commit calls service with collected values, abandon leaves task standalone, commit-error displayed.

## 5. Page wiring — entry points

- [x] 5.1 Task list: add a Convert to Project keybinding + action, available only when the selected task is standalone; push the convert-to-project wizard.
- [x] 5.2 Project view: add Link Task keybinding + action, disabled when no standalone open task exists (query TaskService for candidate existence); when enabled, push the task picker. Handle the picker's `selectedMsg` by calling `ProjectService.LinkTaskToProject`, refresh the task list, and surface any error in the project view.
- [x] 5.3 Project view: add Convert to Task keybinding + action (with confirm), available only when the project is open and has zero tasks; call `ConvertProjectToTask` and navigate away from the now-deleted project.
- [x] 5.4 Project list: add Convert to Task keybinding + action (with confirm), same guards as 5.3; remove the project from the list and reflect the new standalone task.
- [x] 5.5 Centralize the guard predicates (`task standalone`, `project open && zero tasks`) in shared helpers so the three pages do not re-derive them.
- [x] 5.6 Update keymap help/Keys() groups on each affected page.

## 6. Docs

- [x] 6.1 Run `/update-readme` to document the three actions, their entry points, and keybindings; note the someday-link visibility side effect.
- [x] 6.2 Update CLAUDE.md if the service surface or TUI structure description changed. (No CLAUDE.md exists in the project — nothing to update.)
