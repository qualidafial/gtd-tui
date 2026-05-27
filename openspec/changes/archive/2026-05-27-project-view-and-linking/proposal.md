## Why

Tasks have a `ProjectID` foreign key but there is no TUI surface for linking them. Users cannot assign a task to a project, view a project's tasks, or create tasks scoped to a project. This is the second incremental change in the project TUI scope (after the project list tab), enabling the core GTD workflow of organizing next actions under multi-step projects.

## What Changes

### Project picker overlay

- `p` from the task list (or future task view) pushes a project picker overlay
- `huh.Select` over open projects plus "(none)" to unlink, sized to the overlay
- On confirm, updates the task's `ProjectID` via `TaskService.UpdateTask` itself, then dismisses (self-contained, matches taskstatus pattern)
- Lives in `tui/pages/projects/projectpicker/`; tasklist receives a `func(task) screen.Screen` factory from its caller so task packages remain project-unaware

### Project view screen

- Enter on a project in the project list pushes a project view screen onto the overlay stack
- Header displays non-empty project attributes: title, status, outcome, due
- Below the header, an embedded `tasklist.Model` shows all tasks in the project (no status filter by default)
- Task list is scoped via a `TaskService` wrapper that injects `ProjectID` on `ListTasks` and `CreateTask`, delegating everything else to the inner service
- `+`/`insert` creates a new task pre-populated with the project's ID (same key as task list — `n` is retired in favor of `+`/`insert` for all create actions)
- All existing task interactions work within the project context (complete, drop, reopen, reorder, edit, query)
- Esc dismisses back to the project list

### Not in scope

- Project field in taskedit (project assignment happens via picker overlay, not the edit form)
- Project chip on task list rows (deferred)
- Inline "create new project" from picker (deferred)
- Project edit overlay (separate change)
- Project query filter (separate change)

## Known Concerns

### Init-on-return overwrites dirty editor state

The overlay stack calls `Init()` on the parent screen when a child overlay is dismissed (`app.go` line 66). This is correct for list screens that need to refresh data, but would destroy unsaved form state if an overlay is ever pushed on top of an editor. This doesn't affect the current change (picker overlays sit on top of lists, not editors), but must be addressed before overlays can be stacked on editors.

## Capabilities

### New Capabilities
- `project-view-screen`: Project detail view with header and embedded scoped task list
- `project-task-service-wrapper`: TaskService decorator that scopes all operations to a project
- `project-picker-overlay`: Standalone overlay for assigning/unlinking a task's project

### Modified Capabilities
- `project-list-ui`: Enter key on project pushes project view screen
- `task-list-ui`: `p` key pushes project picker overlay for selected task

## Impact

- **tui/pages/projects/projectview/model.go** (new): Project view screen with header + embedded tasklist
- **tui/pages/projects/projectlist.go**: Enter key handler pushes project view
- **service/projecttask.go** or similar (new): `TaskService` wrapper injecting `ProjectID`
- **project picker overlay** (new, location TBD): Select over open projects, updates task on confirm
- **tui/pages/tasks/tasklist/model.go**: `p` key pushes project picker (via injected factory); change create key from `n` to `+`/`insert`
- **tui/pages/tasks/tasklist/keymap.go**: Add `p` binding, rename `n` → `+`/`insert`
- **tui/pages/projects/projectlist.go**: Change create key from `n` to `+`/`insert`
- **tui/pages/projects/projectpicker/model.go** (new): Project picker overlay
