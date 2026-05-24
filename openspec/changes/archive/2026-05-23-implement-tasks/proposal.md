## Why

The GTD TUI has a Task implementation, but it uses the wrong domain model: status encodes inbox/active/waiting/deferred states that belong to separate entities or fields per DESIGN.md. This change corrects the Task domain to match the spec — a breaking revision with no data migration needed (no valuable data exists yet).

## What Changes

- Replace shipped status constants (inbox/active/waiting/deferred) with correct model: Status = pending/done/dropped only
- Add Kind type (next_action/delegated) and Assignee field to Task
- Update 0001_tasks.sql in-place: correct status CHECK constraint, add kind/assignee columns
- Remove fake inbox page from TUI (inbox-as-task-status gone; real inbox comes in implement-inbox)
- Update TUI task pages (taskedit, tasklist, app.go) to use Kind + Assignee instead of old status values
- DeferUntil field already exists; deferred task filtering in default views remains

## Capabilities

### New Capabilities

- `task-entity`: Task domain type with correct fields (Kind, Status=pending/done/dropped, Assignee, Due, DeferUntil) and validation rules
- `task-service`: TaskService interface with Create, Update, List, Get, and status transition methods (CompleteTask, DropTask, ReopenTask)
- `task-sqlite`: SQLite implementation with corrected schema and constraints

### Modified Capabilities

- `tui-tasks`: Task pages updated to reflect Kind/Assignee instead of old status values; fake inbox page removed

## Impact

- Root package: `task.go` — remove old status constants, add TaskKind type and Assignee field
- `sqlite/migrations/0001_tasks.sql`: update in-place with correct status enum and new columns (no data migration; no valuable data)
- `sqlite/task.go`: update column list, scan, and queries
- `service/task.go`: remove references to removed statuses
- TUI: `tui/pages/tasks/taskedit/model.go`, `tui/pages/tasks/tasklist/model.go`, `tui/app.go` — remove inbox page, update status/kind handling
- Tests: update to use correct status values and new fields
