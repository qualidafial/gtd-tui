## 1. Domain Types

- [ ] 1.1 Define TaskKind type and constants (TaskKindNextAction, TaskKindDelegated) in task.go
- [ ] 1.2 Define TaskStatus type and constants (TaskStatusPending, TaskStatusDone, TaskStatusDropped) in task.go
- [ ] 1.3 Define Task struct with all fields (ID, Title, Description, Kind, Status, Assignee, Due, DeferUntil, ProjectID, CreatedAt, UpdatedAt) in task.go
- [ ] 1.4 Define TaskListOptions struct with filter fields (Status, Kind, ProjectID, IncludeDeferred) in task.go

## 2. Service Interface

- [ ] 2.1 Define TaskService interface with CreateTask method in task.go
- [ ] 2.2 Add UpdateTask method to TaskService interface
- [ ] 2.3 Add GetTask method to TaskService interface
- [ ] 2.4 Add ListTasks method to TaskService interface
- [ ] 2.5 Add CompleteTask method to TaskService interface
- [ ] 2.6 Add DropTask method to TaskService interface
- [ ] 2.7 Add ReopenTask method to TaskService interface

## 3. SQLite Migration

- [ ] 3.1 Create sqlite/migrations/0002_tasks.sql with tasks table schema
- [ ] 3.2 Add CHECK constraint for kind enum (next_action, delegated)
- [ ] 3.3 Add CHECK constraint for status enum (pending, done, dropped)
- [ ] 3.4 Add CHECK constraint for non-empty title
- [ ] 3.5 Add foreign key constraint for project_id referencing projects(id)

## 4. SQLite Store Implementation

- [ ] 4.1 Create sqlite/task.go with TaskStore struct
- [ ] 4.2 Implement CreateTask with transaction and timestamp handling
- [ ] 4.3 Implement GetTask with squirrel query builder
- [ ] 4.4 Implement ListTasks with filter options and deferred task filtering
- [ ] 4.5 Implement UpdateTask with comment creation (when provided) in transaction
- [ ] 4.6 Implement CompleteTask with status validation and comment support
- [ ] 4.7 Implement DropTask with status validation and comment support
- [ ] 4.8 Implement ReopenTask with status validation and comment support

## 5. Tests

- [ ] 5.1 Add table-driven tests for CreateTask (success, with project, validation errors)
- [ ] 5.2 Add table-driven tests for GetTask (existing, non-existent)
- [ ] 5.3 Add table-driven tests for ListTasks (by status, by kind, by project, deferred filtering)
- [ ] 5.4 Add table-driven tests for UpdateTask (success, with comment, status change rejection)
- [ ] 5.5 Add table-driven tests for CompleteTask (success, non-pending fails)
- [ ] 5.6 Add table-driven tests for DropTask (success, non-pending fails)
- [ ] 5.7 Add table-driven tests for ReopenTask (from done, from dropped, pending fails)
