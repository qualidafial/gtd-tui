## Why

The GTD methodology depends on capturing *why* decisions were made, not just *what* changed. Currently, when a task status changes or a project is updated, the timeline shows the metadata change but loses the context that motivated it. Comments enable recording reasons alongside changes ("marked done because customer confirmed receipt") and standalone context notes ("blocked on infra ticket") without requiring a metadata edit.

## What Changes

- Add Comment entity with body text, timestamps, and exactly one FK to Task or Project (CHECK constraint)
- Extend UpdateTask and UpdateProject service methods to accept an optional comment string that creates a Comment atomically with the update
- Add CommentService for standalone comment CRUD operations
- Add comments table with appropriate constraints and indexes
- Enable editing of existing comments for corrections

## Capabilities

### New Capabilities
- `comment-entity`: Domain type, service interface, and SQLite implementation for Comment entity with dual-FK constraint (exactly one of TaskID/ProjectID set)
- `edit-with-comment`: Transactional update-plus-comment flow for UpdateTask and UpdateProject

### Modified Capabilities
(none - existing specs remain unchanged; this adds new behavior without modifying existing requirements)

## Impact

- **Domain layer**: New Comment struct and CommentService interface in root package
- **SQLite layer**: New comments table migration, comment.go implementation
- **Service layer**: Updated task.go and project.go to handle comment parameter
- **TUI layer**: (future change) Edit views will need UI for entering comments; comment lists on timeline views
