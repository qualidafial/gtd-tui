## ADDED Requirements

### Requirement: UpdateTask accepts optional comment
The system SHALL extend UpdateTask to accept an optional comment string parameter. When provided, the update operation SHALL create a Comment attached to the Task in the same transaction as the Task update.

#### Scenario: Update task without comment
- **WHEN** UpdateTask is called with empty comment string
- **THEN** Task is updated
- **AND** no Comment is created

#### Scenario: Update task with comment
- **WHEN** UpdateTask is called with non-empty comment string
- **THEN** Task is updated
- **AND** Comment is created with the provided text attached to the Task
- **AND** both operations occur in one transaction

#### Scenario: Update task with comment fails atomically
- **WHEN** UpdateTask with comment fails during Comment creation
- **THEN** Task update is rolled back
- **AND** no partial changes persist

### Requirement: UpdateProject accepts optional comment
The system SHALL extend UpdateProject to accept an optional comment string parameter. When provided, the update operation SHALL create a Comment attached to the Project in the same transaction as the Project update.

#### Scenario: Update project without comment
- **WHEN** UpdateProject is called with empty comment string
- **THEN** Project is updated
- **AND** no Comment is created

#### Scenario: Update project with comment
- **WHEN** UpdateProject is called with non-empty comment string
- **THEN** Project is updated
- **AND** Comment is created with the provided text attached to the Project
- **AND** both operations occur in one transaction

#### Scenario: Update project with comment fails atomically
- **WHEN** UpdateProject with comment fails during Comment creation
- **THEN** Project update is rolled back
- **AND** no partial changes persist

### Requirement: Status transition methods accept optional comment
The system SHALL extend status transition methods (CompleteTask, DropTask, ReopenTask, CompleteProject, DropProject, ParkProject, ReopenProject) to accept an optional comment string. This allows recording the reason for status changes.

#### Scenario: Complete task with reason
- **WHEN** CompleteTask is called with comment "Customer confirmed receipt"
- **THEN** Task status is set to done
- **AND** Comment is created with "Customer confirmed receipt"

#### Scenario: Drop task with reason
- **WHEN** DropTask is called with comment "No longer relevant"
- **THEN** Task status is set to dropped
- **AND** Comment is created with "No longer relevant"

#### Scenario: Complete project with reason
- **WHEN** CompleteProject is called with comment "All deliverables shipped"
- **THEN** Project status is set to done
- **AND** Comment is created with "All deliverables shipped"

### Requirement: Edit-with-comment uses service transaction pattern
The edit-with-comment operation SHALL follow the existing service-level transaction pattern: open a transaction at the service method level, pass it to all internal helpers, and commit or rollback atomically.

#### Scenario: Transaction opened at service level
- **WHEN** UpdateTask with comment is called
- **THEN** service opens transaction before any writes

#### Scenario: Transaction commits on success
- **WHEN** both Task update and Comment creation succeed
- **THEN** transaction commits atomically

#### Scenario: Transaction rolls back on failure
- **WHEN** any operation fails
- **THEN** transaction rolls back all changes

### Requirement: Standalone comment creation
The system SHALL support creating Comments without an associated entity update via CommentService.CreateComment. This enables adding context notes that are not tied to metadata changes (e.g., "blocked on infra ticket", "saw this come up again today").

#### Scenario: Create standalone comment on task
- **WHEN** CreateComment is called with TaskID and body
- **THEN** Comment is created without modifying the Task

#### Scenario: Create standalone comment on project
- **WHEN** CreateComment is called with ProjectID and body
- **THEN** Comment is created without modifying the Project

### Requirement: Comment editing for corrections
The system SHALL support editing existing Comments via CommentService.UpdateComment. Only the body field is editable; the parent reference (TaskID/ProjectID) is immutable after creation.

#### Scenario: Edit comment body
- **WHEN** UpdateComment is called with modified body
- **THEN** Comment body is updated
- **AND** UpdatedAt is refreshed

#### Scenario: Cannot change comment parent
- **WHEN** UpdateComment is called with different TaskID or ProjectID
- **THEN** system ignores the changed parent reference
- **AND** original parent association is preserved
