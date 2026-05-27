## Why

The TUI currently shows only a Tasks tab. Projects exist in the domain, service, and storage layers but have no UI surface. Users cannot view, reorder, or transition projects without direct database access. Adding a Projects tab is the first step toward full project management in the TUI.

## What Changes

- Add a "Projects" tab to the tab container, alongside the existing "Tasks" tab
- Implement a project list screen (`tui/pages/projects/projectlist.go`) displaying all projects with status markers, title, task progress chip (e.g. "3/5"), and due chip
- Minimal project creation via "n" key (title-only inline input, creates open project)
- Support status transitions from the project list: space (complete/reopen), delete (drop), s (park)
- Support reordering open/someday projects with shift+up/shift+down
- Uncomment and wire `ProjectService` into the app (service, main)
- Add a `CountTasksByProject` query to support the task progress chip without loading full task lists

## Capabilities

### New Capabilities
- `project-list-presentation`: Project list rendering with status markers, title, task-progress and due chips, and row selection
- `project-list-ui`: Project list tab screen with quick-create, status transitions (complete/drop/park/reopen), reordering, and data loading

### Modified Capabilities
- `tui-application`: Root model accepts ProjectService; tab container includes Projects tab

## Impact

- **tui/app.go**: Constructor accepts `gtd.ProjectService`; passes to project list screen; adds Projects tab
- **cmd/gtd/main.go**: Uncomment ProjectService creation and wiring
- **service/project.go**: Uncomment ProjectService implementation
- **tui/pages/projects/projectlist.go**: New project list model (replaces stub)
- **sqlite/project.go** or **sqlite/task.go**: Add `CountTasksByProject` query (returns pending/total counts per project ID)
- **project.go** (domain): Add `ProjectTaskCounts` type for the progress chip