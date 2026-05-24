## 1. Domain Types

- [x] 1.1 Define TaskKind type and constants (TaskKindNextAction, TaskKindDelegated) in task.go
- [x] 1.2 Replace shipped status constants with correct set: TaskStatusPending, TaskStatusDone, TaskStatusDropped; remove TaskStatusInbox, TaskStatusActive, TaskStatusWaiting, TaskStatusDeferred
- [x] 1.3 Add Kind TaskKind and Assignee string fields to Task struct; remove ProjectID (added in implement-projects)
- [x] 1.4 Update TaskFilter: remove Status values that no longer exist; add Kind filter field

## 2. Service Interface

- [x] 2.1 Define TaskService interface with CreateTask method in task.go
- [x] 2.2 Add UpdateTask method to TaskService interface
- [x] 2.3 Add GetTask method to TaskService interface
- [x] 2.4 Add ListTasks method to TaskService interface
- [x] 2.5 Add CompleteTask method to TaskService interface
- [x] 2.6 Add DropTask method to TaskService interface
- [x] 2.7 Add ReopenTask method to TaskService interface

## 3. SQLite Migration

- [x] 3.1 Consolidate into sqlite/migrations/0001_tasks.sql: replace status enum (pending/done/dropped), add kind/assignee/order_key columns and the order_key index; delete 0002_task_order_key.sql
- [x] 3.2 Verify CHECK constraint for kind enum (next_action, delegated)
- [x] 3.3 Verify CHECK constraint for status enum (pending, done, dropped)
- [x] 3.4 Verify CHECK constraint for non-empty title

## 4. SQLite Store Implementation

- [x] 4.1 Create sqlite/task.go with TaskStore struct
- [x] 4.2 Implement CreateTask with transaction and timestamp handling
- [x] 4.3 Implement GetTask with squirrel query builder
- [x] 4.4 Implement ListTasks with filter options and deferred task filtering
- [x] 4.5 Implement UpdateTask (comment support added in implement-comments)
- [x] 4.6 Implement CompleteTask with status validation (comment param added in implement-comments)
- [x] 4.7 Implement DropTask with status validation (comment param added in implement-comments)
- [x] 4.8 Implement ReopenTask with status validation (comment param added in implement-comments)

## 5. TUI Updates

- [x] 5.1 Remove fake inbox tasklist page from tui/app.go (TaskStatusInbox filter gone)
- [x] 5.2 Replace status dropdown in taskedit with Kind selector (next_action/delegated) and Assignee input field
- [x] 5.3 Update tasklist item rendering to use Kind instead of old status values
- [x] 5.4 Update all TaskFilter usage in TUI to remove references to removed statuses

## 6. Tests

- [x] 6.1 Update existing table-driven tests to use TaskStatusPending instead of removed statuses
- [x] 6.2 Add tests for CreateTask with Kind=delegated requires non-empty Assignee
- [x] 6.3 Add tests for ListTasks with Kind filter
- [x] 6.4 Add tests for CompleteTask (success from pending, non-pending fails)
- [x] 6.5 Add tests for DropTask (success from pending, non-pending fails)
- [x] 6.6 Add tests for ReopenTask (from done, from dropped, pending fails)
- [x] 6.7 Add tests for deferred task filtering (DeferUntil in future excluded from default list)
