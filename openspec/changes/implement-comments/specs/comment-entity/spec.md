## ADDED Requirements

### Requirement: Comment entity structure
The system SHALL provide a Comment entity representing short, event-shaped text attached to exactly one Task or Project. A Comment SHALL have:
- ID (int64, auto-assigned)
- Body (non-empty string)
- TaskID (nullable int64 FK to tasks)
- ProjectID (nullable int64 FK to projects)
- CreatedAt (timestamp, auto-assigned)
- UpdatedAt (timestamp, auto-maintained)

Exactly one of TaskID or ProjectID SHALL be set (enforced by CHECK constraint). Comments use value semantics (no pointers in service interfaces).

#### Scenario: Comment on task has TaskID set
- **WHEN** a Comment is created for a Task
- **THEN** TaskID references the Task
- **AND** ProjectID is nil

#### Scenario: Comment on project has ProjectID set
- **WHEN** a Comment is created for a Project
- **THEN** ProjectID references the Project
- **AND** TaskID is nil

#### Scenario: Exactly one FK constraint enforced
- **WHEN** attempting to create a Comment with both TaskID and ProjectID set
- **THEN** database rejects with CHECK constraint violation

#### Scenario: Exactly one FK constraint prevents both nil
- **WHEN** attempting to create a Comment with both TaskID and ProjectID nil
- **THEN** database rejects with CHECK constraint violation

### Requirement: Comment body validation
The system SHALL require a non-empty body for all Comments. Empty or whitespace-only bodies SHALL be rejected.

#### Scenario: Non-empty body required
- **WHEN** attempting to create a Comment with empty body
- **THEN** system rejects with validation error

#### Scenario: Whitespace-only body rejected
- **WHEN** attempting to create a Comment with whitespace-only body
- **THEN** system rejects with validation error

### Requirement: CommentService interface
The system SHALL provide a CommentService interface in the root package with the following methods:
- Comment(ctx, id) returns a single Comment by ID
- Comments(ctx, filter) returns Comments matching filter criteria
- CreateComment(ctx, Comment) returns created Comment with server-assigned fields
- UpdateComment(ctx, Comment) returns updated Comment
- DeleteComment(ctx, id) removes the Comment

#### Scenario: Create comment returns populated value
- **WHEN** CreateComment is called
- **THEN** returned Comment has ID, CreatedAt, UpdatedAt populated

#### Scenario: Update comment refreshes timestamp
- **WHEN** UpdateComment is called
- **THEN** returned Comment has UpdatedAt refreshed

#### Scenario: Get comment by ID
- **WHEN** Comment is called with valid ID
- **THEN** system returns the matching Comment

#### Scenario: Get comment not found
- **WHEN** Comment is called with non-existent ID
- **THEN** system returns not found error

### Requirement: CommentFilter for querying comments
The system SHALL provide a CommentFilter struct supporting filtering by:
- TaskID (optional int64) - comments on a specific task
- ProjectID (optional int64) - comments on a specific project
- TaskIDs (optional []int64) - comments on any of the specified tasks
- ProjectIDs (optional []int64) - comments on any of the specified projects

#### Scenario: Filter comments by task
- **WHEN** Comments is called with TaskID filter
- **THEN** system returns only Comments for that Task

#### Scenario: Filter comments by project
- **WHEN** Comments is called with ProjectID filter
- **THEN** system returns only Comments for that Project

#### Scenario: Bulk load comments for multiple tasks
- **WHEN** Comments is called with TaskIDs filter
- **THEN** system returns Comments for all specified Tasks in a single query

### Requirement: Comments ordered by creation time
The system SHALL return Comments ordered by CreatedAt ascending (oldest first) by default. This supports chronological timeline display.

#### Scenario: Comments returned in chronological order
- **WHEN** Comments is called for a Task
- **THEN** results are ordered by CreatedAt ascending

### Requirement: Comment cascade delete
The system SHALL delete Comments when their parent Task or Project is deleted. This is enforced via ON DELETE CASCADE on the foreign keys.

#### Scenario: Delete task cascades to comments
- **WHEN** a Task is deleted
- **THEN** all Comments with that TaskID are deleted

#### Scenario: Delete project cascades to comments
- **WHEN** a Project is deleted
- **THEN** all Comments with that ProjectID are deleted

### Requirement: Comments table schema
The system SHALL store Comments in a `comments` table with:
- INTEGER PRIMARY KEY for id (auto-increment)
- TEXT NOT NULL for body with CHECK(body != '')
- INTEGER for task_id with REFERENCES tasks(id) ON DELETE CASCADE
- INTEGER for project_id with REFERENCES projects(id) ON DELETE CASCADE
- TEXT NOT NULL for created_at (ISO8601 UTC)
- TEXT NOT NULL for updated_at (ISO8601 UTC)
- CHECK constraint enforcing exactly one of task_id/project_id is set
- Index on task_id for efficient task timeline queries
- Index on project_id for efficient project timeline queries

#### Scenario: Insert comment for task
- **WHEN** inserting a row with task_id set and project_id null
- **THEN** row is created successfully

#### Scenario: Insert comment for project
- **WHEN** inserting a row with project_id set and task_id null
- **THEN** row is created successfully
