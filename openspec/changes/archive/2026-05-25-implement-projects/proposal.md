## Why

The GTD methodology requires multi-step outcomes (Projects) as a core organizational unit. Without Projects, users cannot group related tasks, track progress toward outcomes, or properly implement the Organize and Engage workflows. The foundation specs define the Project entity and its relationship to Tasks, but the implementation does not yet exist.

## What Changes

- Add Project domain type with title, outcome, description, due, and status (open/someday/done/dropped)
- Add ProjectService interface with CRUD operations, status transitions, and reordering
- Add SQLite implementation for Project persistence
- Implement status transition methods:
  - CompleteProject: marks project done, cascades or detaches pending tasks based on flag
  - DropProject: marks project dropped, cascades or detaches pending tasks based on flag
  - ParkProject: sets status to someday with a fresh order key in the someday ordering, filters tasks from default views (no task status change)
  - ReopenProject: restores a someday/done/dropped project to open with a fresh order key (mirrors ReopenTask; no task status change)
- Enforce invariant: no pending tasks under closed (done/dropped) projects
- Add ProjectID *int64 to Task struct and a new migration adding project_id column to tasks table (FK to projects with ON DELETE SET NULL)
- Add fractional-indexed ordering for open and someday projects (order_key column, MoveProjectUp/MoveProjectDown), mirroring the task ordering pattern. Open and someday projects are ordered independently within their status groups; done/dropped projects are unordered.
- ListProjects returns projects in three tiers: open (by order_key), someday (by order_key), done/dropped (by status_changed_at DESC)
- Tasks under someday projects are excluded by default (IncludeSomedayProjects filter field, default false)
- Follow-on for `task-query-filter`: once tasks link to projects, widen the task-list free-text search to also match the linked project's name (JOIN projects in the sqlite `ListTasks` free-text clause). The query grammar needs no change — only the sqlite free-text clause widens. See task-query-filter design.md "Future Considerations".

## Out of Scope

- **Comments on projects.** The Comment entity does not exist yet. Project service methods (UpdateProject, CompleteProject, DropProject, ParkProject, ReopenProject) ship without comment parameters. `implement-comments` owns adding comment support — it will re-break these signatures to add an optional comment string and atomic Comment creation. API churn is acceptable; see implement-comments `edit-with-comment` spec.

## Capabilities

### New Capabilities
- `project-entity`: Project domain type with fields and validation
- `project-service`: ProjectService interface for CRUD, status transitions, and reordering
- `project-task-relationship`: Task to Project linking and cascade behaviors

### Modified Capabilities

## Impact

- Root package: new Project type and ProjectService interface in project.go; Task struct gains ProjectID *int64
- sqlite/: new project.go, new migration for projects table (with order_key) and project_id column on tasks
- sqlite/task.go: add ProjectID to column list and scan; query support for filtering by ProjectID and project status (IncludeSomedayProjects); widen free-text search to match linked project name (task-query-filter follow-on)
- Tests: new project_test.go for service tests including ordering; update task tests for ProjectID