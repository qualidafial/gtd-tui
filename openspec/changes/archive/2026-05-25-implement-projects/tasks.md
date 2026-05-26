## 1. Domain Types

- [x] 1.1 Rewrite Project type in project.go: status set open/someday/done/dropped
- [x] 1.2 Add ProjectStatus constants (open, someday, done, dropped)
- [x] 1.3 Add ProjectFilter type with Status filter field
- [x] 1.4 Add ProjectService interface with CRUD methods (GetProject, ListProjects, CreateProject, UpdateProject) — no comment params (deferred to implement-comments)
- [x] 1.5 Add ProjectService transition methods: CompleteProject(ctx, id, cascade, at), DropProject(ctx, id, cascade, at), ParkProject(ctx, id, at), ReopenProject(ctx, id, at)
- [x] 1.6 Add ProjectService reordering methods: MoveProjectUp(ctx, id), MoveProjectDown(ctx, id)
- [x] 1.7 Re-add ProjectID *int64 field to Task type (removed in 05-24 reconciliation) for direct FK relationship
- [x] 1.8 Remove project_task.go (join table approach not used per DESIGN.md)

## 2. Database Schema

- [x] 2.1 Create migration 0003_projects.sql with projects table, status CHECK constraint, order_key, and status_changed_at column (0002 already taken by task_status_changed_at)
- [x] 2.2 Add project_id column to tasks table with FK to projects(id) ON DELETE SET NULL
- [x] 2.3 Add index on tasks.project_id for reverse queries
- [x] 2.4 Remove future.sql scaffolding (project_tasks join table not needed)

## 3. SQLite Implementation - Basic CRUD

- [x] 3.1 Rewrite sqlite/project.go: no DeleteProject, corrected status values
- [x] 3.2 Update CreateProject to default status to open if not specified, seed status_changed_at to created_at, assign order_key for open and someday projects
- [x] 3.3 Implement UpdateProject (no comment param; comment support added later by implement-comments)
- [x] 3.4 Rename Project method to GetProject for consistency with specs
- [x] 3.5 Add scanProject function with proper nullable field handling (include status_changed_at)
- [x] 3.6 Update sqlite/task.go to include project_id in taskColumns and scanTask

## 4. SQLite Implementation - Status Transitions

- [x] 4.1 Implement CompleteProject(id, cascade, at) with transaction (set project status_changed_at = at, clear order_key)
- [x] 4.2 Implement DropProject(id, cascade, at) with transaction (set project status_changed_at = at, clear order_key)
- [x] 4.3 Implement ParkProject(id, at) with transaction (set project status_changed_at = at, assign fresh order_key in someday ordering)
- [x] 4.4 Implement ReopenProject(id, at) with transaction: someday/done/dropped → open, set status_changed_at = at, assign fresh order_key, tasks untouched
- [x] 4.5 Add helper to cascade status to pending tasks (mark done/dropped, set StatusChangedAt = at)
- [x] 4.6 Add helper to detach pending tasks (set project_id = NULL)

## 5. Task Filtering by Project Status

- [x] 5.1 Add ProjectID filter field to TaskFilter
- [x] 5.2 Add IncludeSomedayProjects bool filter field to TaskFilter (default false excludes someday-project tasks)
- [x] 5.3 Update ListTasks query to LEFT JOIN projects and exclude someday-project tasks by default (when !IncludeSomedayProjects)
- [x] 5.4 Update ListTasks query to filter by ProjectID when specified

## 6. Project Ordering

- [x] 6.1 Add order_key column to projects table in migration
- [x] 6.2 Assign order_key on CreateProject for open and someday projects; NULL for done/dropped
- [x] 6.3 Assign fresh order_key on transitions to open or someday; clear on transitions to done/dropped
- [x] 6.4 Sort ListProjects in 3 tiers: open (by order_key), someday (by order_key), done/dropped (by status_changed_at DESC)
- [x] 6.5 Implement MoveProjectUp/MoveProjectDown with fractional-index shift + renumber fallback
- [x] 6.6 Reject reordering for done/dropped projects only; open and someday are independently orderable

## 7. Tests - Project CRUD

- [x] 7.1 Rewrite sqlite/project_test.go with corrected status values
- [x] 7.2 Add test for CreateProject with default open status
- [x] 7.3 Add test for CreateProject validation (empty title rejected)
- [x] 7.4 Add test for CreateProject validation (invalid status rejected)
- [x] 7.5 Add test for GetProject returns error for non-existent ID
- [x] 7.6 Add test for ListProjects by status filter
- [x] 7.7 Add test for UpdateProject refreshes UpdatedAt

## 8. Tests - Status Transitions

- [x] 8.1 Add test for CompleteProject with cascade=true marks pending tasks done
- [x] 8.2 Add test for CompleteProject with cascade=false detaches pending tasks
- [x] 8.3 Add test for CompleteProject preserves done/dropped tasks on project
- [x] 8.4 Add test for DropProject with cascade=true marks pending tasks dropped
- [x] 8.5 Add test for DropProject with cascade=false detaches pending tasks
- [x] 8.6 Add test for ParkProject sets status to someday without changing tasks
- [x] 8.7 Add test for ReopenProject restores open from someday, done, and dropped without changing task statuses
- [x] 8.8 Add test for CompleteProject/DropProject cascade sets each task's StatusChangedAt to the supplied instant
- [x] 8.9 Add test that every transition sets the project's StatusChangedAt to the supplied instant; CreateProject seeds it to created_at; UpdateProject leaves it unchanged

## 9. Tests - Task-Project Relationship

- [x] 9.1 Add test for task with ProjectID references valid project
- [x] 9.2 Add test for default filter excludes someday-project tasks; IncludeSomedayProjects: true includes them
- [x] 9.3 Add test for ListTasks with ProjectID filter returns only project tasks
- [x] 9.4 Add test for pending task under closed project is visible (invariant check)

## 10. Tests - Project Ordering

- [x] 10.1 Add test for 3-tier ordering: open by order_key, someday by order_key, done/dropped by status_changed_at DESC
- [x] 10.2 Add test for open projects preserve creation order
- [x] 10.3 Add test for MoveProjectUp shifts project earlier
- [x] 10.4 Add test for MoveProjectDown shifts project later
- [x] 10.5 Add test for MoveProjectUp/Down rejects done/dropped projects
- [x] 10.6 Add test for someday projects order independently from open projects
- [x] 10.7 Add test for ParkProject assigns order_key (appears after existing someday projects)

## 11. Cleanup

- [x] 11.1 Remove sqlite/project_task.go (join table not used)
- [x] 11.2 Run go mod tidy to clean up dependencies
- [x] 11.3 Run all tests to verify implementation