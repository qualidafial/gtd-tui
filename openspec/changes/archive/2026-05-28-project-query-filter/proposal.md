## Why

The project list has no filtering — all projects display in a fixed order with no way to narrow the view. Task lists already have a query bar supporting `status:`, free-text search, and date predicates. Projects need the same capability. Additionally, the task list query bar currently reserves 3 lines even when unfocused (prompt + underline + error), wasting vertical space on every screen that embeds a task list.

Error handling is also inconsistent: each screen manages its own `err` field, renders errors in different locations (inline footer, beneath forms, replacing the view), and handles esc-to-clear independently. This should be centralized in `app.Model` — errors display uniformly in the help bar row, and esc-to-clear is handled once at the app level.

## What Changes

- Add a `projectquery` parser analogous to `taskquery`, supporting `status:` and free-text search on project title/outcome. No date predicates initially (projects have only `due`, no `defer`/`ready`).
- Expand `ProjectFilter` to support `Search []string` for free-text matching.
- Add a query bar to the project list (`projectlist.go`), with `/` to focus, enter to apply, esc to cancel — same interaction pattern as task list.
- Extract the query bar widget from inline task list code into a reusable `querybar` component, so both project list and task list share one implementation.
- Make the query bar always one line. When a parse error exists, highlight the offending range inline using color and underline styling applied to the textinput's rendered output (replacing the current `^^^` underline on a second row). The error message itself displays in the app error bar.
- Centralize error handling in `tui.Model`: screens return `error` as a tea.Msg, `tui.Model` catches it, displays the error in the help bar row (replacing the help text), and handles esc to clear. Remove per-screen `err` fields and error rendering for ambient errors (projectlist, projectpicker). Save-error overlays (taskedit, projectedit, projectstatus, taskstatus) keep internal error state to block form re-fire, but display moves to the help bar.

## Capabilities

### New Capabilities
- `project-query`: Parser for project list query strings into `ProjectFilter`, supporting `status:` and free-text search tokens.
- `query-bar`: Reusable single-line TUI query bar component with focus/blur, inline error highlighting (color + underline on offending range), and debounced validation.
- `app-error-bar`: Centralized error display in the app help bar row — screens return `error` messages, app.Model catches them, renders error text in place of help, esc clears.

### Modified Capabilities
- `project-list-ui`: Add query bar with `/` key to focus, enter to apply filter, esc to cancel. Remove per-screen error display (errors go to app error bar).
- `task-list-query-ui`: Replace inline query bar implementation with the shared `query-bar` component. Query bar collapses to one line when unfocused.

## Impact

- New package: `internal/projectquery/`
- New package: `tui/components/querybar/`
- Modified: `project.go` (`ProjectFilter` gains `Search` field)
- Modified: `sqlite/project.go` (free-text search in `ListProjects`)
- Modified: `tui/app.go` (error interception, help bar replacement, esc-to-clear)
- Modified: `tui/pages/projects/projectlist.go` (query bar integration, remove `err` field and error rendering)
- Modified: `tui/pages/projects/keymap.go` (query focus/apply/cancel bindings)
- Modified: `tui/pages/tasks/tasklist/model.go` (extract query bar to shared component)
- Modified: `tui/pages/tasks/tasklist/keymap.go` (query bindings reference shared component)
- Modified: `tui/pages/projects/projectpicker/model.go` (remove per-screen error display)
- Modified: `tui/pages/tasks/taskedit/model.go` (error display moves to help bar, keep internal error state for retry blocking)
- Modified: `tui/pages/projects/projectedit/model.go` (same as taskedit)
- Modified: `tui/pages/projects/projectstatus/model.go` (same)
- Modified: `tui/pages/tasks/taskstatus/model.go` (same)
