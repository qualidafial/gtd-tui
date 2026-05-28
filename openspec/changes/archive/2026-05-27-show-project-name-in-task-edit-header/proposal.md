## Why

When editing a task that belongs to a project, there's no indication of which project it's linked to. The user loses context about where the task fits in their workflow.

## What Changes

- The task edit overlay's read-only header gains a "Project" line showing the project's title when the task has a non-nil ProjectID.
- The taskedit model accepts a project name (resolved by the caller) so it can render this line without taking a ProjectService dependency.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `task-edit-ui`: Add a "Project" line to the read-only header for tasks linked to a project.

## Impact

- `tui/pages/tasks/taskedit/model.go` — accept project name, render new header line
- Callers that construct `taskedit.New(...)` — resolve project name from ProjectID before opening the overlay
