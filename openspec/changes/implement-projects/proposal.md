## Why

The GTD methodology requires multi-step outcomes (Projects) as a core organizational unit. Without Projects, users cannot group related tasks, track progress toward outcomes, or properly implement the Organize and Engage workflows. The foundation specs define the Project entity and its relationship to Tasks, but the implementation does not yet exist.

## What Changes

- Add Project domain type with title, outcome, description, due, and status (active/someday/done/dropped)
- Add ProjectService interface with CRUD operations and status transitions
- Add SQLite implementation for Project persistence
- Implement status transition methods:
  - CompleteProject: marks project done, cascades or detaches pending tasks based on flag
  - DropProject: marks project dropped, cascades or detaches pending tasks based on flag
  - ParkProject: sets status to someday, filters tasks from default views (no task status change)
  - UnparkProject: restores status to active
- Enforce invariant: no pending tasks under closed (done/dropped) projects
- Support Task.ProjectID relationship for task-to-project linking

## Capabilities

### New Capabilities
- `project-entity`: Project domain type with fields and validation
- `project-service`: ProjectService interface for CRUD and status transitions
- `project-task-relationship`: Task to Project linking and cascade behaviors

### Modified Capabilities

## Impact

- Root package: new Project type and ProjectService interface in project.go
- sqlite/: new project.go with SQLite implementation, new migration for projects table
- sqlite/task.go: query support for filtering tasks by project status (someday filtering)
- Tests: new project_test.go for service tests
