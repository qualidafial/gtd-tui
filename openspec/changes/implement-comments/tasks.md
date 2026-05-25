## 1. Domain Layer

- [ ] 1.1 Create Comment struct in root package (comment.go) with ID, Body, TaskID, ProjectID, CreatedAt, UpdatedAt fields using value semantics and pointer types for nullable FKs
- [ ] 1.2 Create CommentService interface with Comment, Comments, CreateComment, UpdateComment, DeleteComment methods
- [ ] 1.3 Create CommentFilter struct with TaskID, ProjectID, TaskIDs, ProjectIDs fields and filter builder methods

## 2. Database Migration

- [ ] 2.1 Create migration file (0003_comments.sql or higher, after projects migration) with comments table schema
- [ ] 2.2 Add CHECK constraint enforcing exactly one of task_id/project_id is set
- [ ] 2.3 Add CHECK constraint for non-empty body
- [ ] 2.4 Add foreign key constraints with ON DELETE CASCADE for task_id and project_id
- [ ] 2.5 Add indexes on task_id and project_id for efficient timeline queries

## 3. SQLite Implementation - CommentService

- [ ] 3.1 Create sqlite/comment.go with commentColumns and scanComment helper following existing patterns
- [ ] 3.2 Implement Comment(ctx, id) method to fetch single comment by ID
- [ ] 3.3 Implement Comments(ctx, filter) method with TaskID/ProjectID/TaskIDs/ProjectIDs filtering and CreatedAt ASC ordering
- [ ] 3.4 Implement CreateComment(ctx, Comment) with server-assigned ID and timestamps
- [ ] 3.5 Implement UpdateComment(ctx, Comment) that updates only body and UpdatedAt, preserving original parent FK
- [ ] 3.6 Implement DeleteComment(ctx, id) method

## 4. Edit-with-Comment - Task Integration

- [ ] 4.1 Modify TaskService.UpdateTask signature to accept comment string parameter
- [ ] 4.2 Update sqlite/task.go UpdateTask to create Comment within existing RunTx when comment is non-empty
- [ ] 4.3 Update all existing UpdateTask call sites to pass empty string for comment parameter
- [ ] 4.4 Add comment parameter to DropTask and other status transition methods

## 5. Edit-with-Comment - Project Integration

- [ ] 5.1 Modify ProjectService.UpdateProject signature to accept comment string parameter (ProjectService exists as of implement-projects)
- [ ] 5.2 Update sqlite/project.go UpdateProject to create Comment within transaction when comment is non-empty
- [ ] 5.3 Add comment parameter to CompleteProject, DropProject, ParkProject, ReopenProject methods

## 6. Tests

- [ ] 6.1 Create sqlite/comment_test.go with table-driven tests using openTestDB helper
- [ ] 6.2 Test CreateComment for task with valid data, verify returned ID and timestamps
- [ ] 6.3 Test CreateComment for project with valid data
- [ ] 6.4 Test CHECK constraint violation when both TaskID and ProjectID are set
- [ ] 6.5 Test CHECK constraint violation when both TaskID and ProjectID are nil
- [ ] 6.6 Test CHECK constraint violation for empty body
- [ ] 6.7 Test Comments filter by single TaskID
- [ ] 6.8 Test Comments filter by TaskIDs (bulk load)
- [ ] 6.9 Test Comments ordering is CreatedAt ASC
- [ ] 6.10 Test UpdateComment preserves parent FK even when caller tries to change it
- [ ] 6.11 Test cascade delete when Task is deleted
- [ ] 6.12 Test UpdateTask with comment creates Comment atomically
- [ ] 6.13 Test UpdateTask with empty comment creates no Comment
- [ ] 6.14 Test UpdateTask with comment rolls back both on failure
