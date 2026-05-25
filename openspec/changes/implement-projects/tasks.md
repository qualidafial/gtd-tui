## 1. Domain Types

- [ ] 1.1 Rewrite Project type in project.go (scaffold is pre-reconciliation): status set active/someday/done/dropped, no "deferred"
- [ ] 1.2 Add ProjectStatus constants (active, someday, done, dropped)
- [ ] 1.3 Add ProjectFilter type with Status filter field
- [ ] 1.4 Add ProjectService interface with CRUD methods (GetProject, ListProjects, CreateProject, UpdateProject) — no comment params (deferred to implement-comments)
- [ ] 1.5 Add ProjectService transition methods: CompleteProject(ctx, id, cascade, at), DropProject(ctx, id, cascade, at), ParkProject(ctx, id, at), ReopenProject(ctx, id, at)
- [ ] 1.6 Re-add ProjectID *int64 field to Task type (removed in 05-24 reconciliation) for direct FK relationship
- [ ] 1.7 Remove project_task.go (join table approach not used per DESIGN.md)

## 2. Database Schema

- [ ] 2.1 Create migration 0003_projects.sql with projects table, status CHECK constraint, and status_changed_at column (0002 already taken by task_status_changed_at)
- [ ] 2.2 Add project_id column to tasks table with FK to projects(id) ON DELETE SET NULL
- [ ] 2.3 Add index on tasks.project_id for reverse queries
- [ ] 2.4 Remove future.sql scaffolding (project_tasks join table not needed)

## 3. SQLite Implementation - Basic CRUD

- [ ] 3.1 Rewrite sqlite/project.go (scaffold is pre-reconciliation): no DeleteProject, corrected status values
- [ ] 3.2 Update CreateProject to default status to active if not specified, and seed status_changed_at to created_at
- [ ] 3.3 Implement UpdateProject (no comment param; comment support added later by implement-comments)
- [ ] 3.4 Rename Project method to GetProject for consistency with specs
- [ ] 3.5 Add scanProject function with proper nullable field handling (include status_changed_at)
- [ ] 3.6 Update sqlite/task.go to include project_id in taskColumns and scanTask

## 4. SQLite Implementation - Status Transitions

- [ ] 4.1 Implement CompleteProject(id, cascade, at) with transaction (set project status_changed_at = at)
- [ ] 4.2 Implement DropProject(id, cascade, at) with transaction (set project status_changed_at = at)
- [ ] 4.3 Implement ParkProject(id, at) with transaction (set project status_changed_at = at)
- [ ] 4.4 Implement ReopenProject(id, at) with transaction: someday/done/dropped → active, set status_changed_at = at, tasks untouched
- [ ] 4.5 Add helper to cascade status to pending tasks (mark done/dropped, set StatusChangedAt = at)
- [ ] 4.6 Add helper to detach pending tasks (set project_id = NULL)

## 5. Task Filtering by Project Status

- [ ] 5.1 Add ProjectID filter field to TaskFilter
- [ ] 5.2 Add ExcludeSomedayProjects bool filter field to TaskFilter
- [ ] 5.3 Update Tasks query to JOIN projects and filter by project status when ExcludeSomedayProjects is true
- [ ] 5.4 Update Tasks query to filter by ProjectID when specified

## 6. Tests - Project CRUD

- [ ] 6.1 Rewrite sqlite/project_test.go (scaffold is pre-reconciliation) with corrected status values
- [ ] 6.2 Add test for CreateProject with default active status
- [ ] 6.3 Add test for CreateProject validation (empty title rejected)
- [ ] 6.4 Add test for CreateProject validation (invalid status rejected)
- [ ] 6.5 Add test for GetProject returns error for non-existent ID
- [ ] 6.6 Add test for ListProjects by status filter
- [ ] 6.7 Add test for UpdateProject refreshes UpdatedAt

## 7. Tests - Status Transitions

- [ ] 7.1 Add test for CompleteProject with cascade=true marks pending tasks done
- [ ] 7.2 Add test for CompleteProject with cascade=false detaches pending tasks
- [ ] 7.3 Add test for CompleteProject preserves done/dropped tasks on project
- [ ] 7.4 Add test for DropProject with cascade=true marks pending tasks dropped
- [ ] 7.5 Add test for DropProject with cascade=false detaches pending tasks
- [ ] 7.6 Add test for ParkProject sets status to someday without changing tasks
- [ ] 7.7 Add test for ReopenProject restores active from someday, done, and dropped without changing task statuses
- [ ] 7.8 Add test for CompleteProject/DropProject cascade sets each task's StatusChangedAt to the supplied instant
- [ ] 7.9 Add test that every transition sets the project's StatusChangedAt to the supplied instant; CreateProject seeds it to created_at; UpdateProject leaves it unchanged

## 8. Tests - Task-Project Relationship

- [ ] 8.1 Add test for task with ProjectID references valid project
- [ ] 8.2 Add test for ListTasks with ExcludeSomedayProjects filters parked project tasks
- [ ] 8.3 Add test for ListTasks with ProjectID filter returns only project tasks
- [ ] 8.4 Add test for pending task under closed project is visible (invariant check)

## 9. Cleanup

- [ ] 9.1 Remove sqlite/project_task.go (join table not used)
- [ ] 9.2 Run go mod tidy to clean up dependencies
- [ ] 9.3 Run all tests to verify implementation
