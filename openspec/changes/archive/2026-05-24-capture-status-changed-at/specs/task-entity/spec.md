## MODIFIED Requirements

### Requirement: Task struct definition
The Task struct SHALL be defined in the root package (`gtd`) with the following fields: ID (int64), Title (string), Description (string), Kind (TaskKind), Status (TaskStatus), Assignee (string), Due (*time.Time), DeferUntil (*time.Time), CreatedAt (time.Time), UpdatedAt (time.Time), StatusChangedAt (time.Time). (The ProjectID field is added by `implement-projects`, not this change.)

#### Scenario: Task struct has all fields
- **WHEN** defining a Task
- **THEN** all fields are present with correct types
- **AND** nullable fields use pointer types
- **AND** StatusChangedAt is a non-pointer time.Time

### Requirement: Task timestamps
CreatedAt, UpdatedAt, and StatusChangedAt SHALL be time.Time values stored as UTC. CreatedAt is set on creation. UpdatedAt is set on creation and updated on every modification (record time). StatusChangedAt is set on creation (equal to CreatedAt, representing the transition into pending) and updated on every status transition to the instant the transition is recorded as having occurred (event time); it is NOT changed by non-status edits.

#### Scenario: Task creation sets timestamps
- **WHEN** a task is created
- **THEN** CreatedAt and UpdatedAt are both set to current UTC time
- **AND** StatusChangedAt is set equal to CreatedAt

#### Scenario: Task update refreshes UpdatedAt only
- **WHEN** a task is updated (a non-status edit)
- **THEN** UpdatedAt is set to current UTC time
- **AND** CreatedAt remains unchanged
- **AND** StatusChangedAt remains unchanged

#### Scenario: Status transition sets StatusChangedAt
- **WHEN** a task's status is transitioned with a given instant
- **THEN** StatusChangedAt is set to that instant (in UTC)
- **AND** UpdatedAt is refreshed to current UTC time
</content>
