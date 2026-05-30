## Why

When the global tasks tab shows a mixed list (e.g., `status:open ready:now`), nothing on a row tells you which project a task belongs to. The user has to open the editor — or recognize the title — to know whether "Email the contractor" is the kitchen remodel or the rental. The data is already loaded (`task.ProjectID`) and the name is already resolvable via the existing `ProjectNameFunc`; the row renderer just doesn't use it.

## What Changes

- Add a per-task chip showing the parent project's title when `task.ProjectID != nil` and a `ProjectNameFunc` is available. Chip renders inline after the title alongside the existing due/defer/assignee chips, in a distinct color.
- Suppress the project chip on the task list inside a project view (every row would be the same project — pure noise). The project view's `tasklist.New` call opts out.
- Existing chip rules (suppression on dropped tasks, etc.) apply.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `task-list-presentation`: add a project chip to the row's chip set, with ordering and visibility rules.

## Impact

- Code: `tui/pages/tasks/tasklist` (render path adds project chip; `New` gains an opt-out for in-project use, or the existing `ProjectNameFunc` is interpreted as "render the chip if non-nil" — design.md picks the path); `tui/pages/projects/projectview/model.go` (suppress the chip).
- APIs: `tasklist.New` signature may change to add an option/flag.
- No DB or service changes. No new resolver call — the existing `projectNameFn` is already invoked when the editor opens; the renderer will call it per visible row.