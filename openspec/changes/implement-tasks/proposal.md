## Why

The GTD TUI needs Task entity implementation to support actionable items - the core unit of work in GTD methodology. Without tasks, users cannot capture next actions, track delegated items, or organize work under projects.

## What Changes

- Add Task domain type with Kind (next_action/delegated), Status (pending/done/dropped), Due, DeferUntil, and ProjectID fields
- Add TaskService interface with CRUD operations and status transition methods
- Add SQLite implementation with CHECK constraints for enum validation
- Add deferred task filtering in default views (tasks with future DeferUntil are hidden)
- Add Assignee field for delegated tasks

## Capabilities

### New Capabilities

- `task-entity`: Task domain type definition with all fields and validation rules
- `task-service`: TaskService interface with Create, Update, List, Get, and status transition methods (CompleteTask, DropTask, ReopenTask)
- `task-sqlite`: SQLite implementation of TaskService with migrations, CHECK constraints, and transactions

### Modified Capabilities

(none - this is new implementation of requirements defined in foundation specs)

## Impact

- Root package: new `task.go` with Task struct and TaskService interface
- sqlite/: new task store implementation and migration file
- Tests: table-driven tests for all service methods
- Foundation specs in `domain-model` and `architecture` define the requirements; this change implements them
