## ADDED Requirements

### Requirement: Project entity fields
The system SHALL provide a Project entity with the following fields:
- ID (int64): unique identifier, server-assigned
- Title (string): short project name for lists, non-empty
- Outcome (string): desired end state statement
- Description (string): detailed project description
- Due (*time.Time): optional firm deadline
- Status (ProjectStatus): one of active, someday, done, dropped
- CreatedAt (time.Time): creation timestamp, server-assigned
- UpdatedAt (time.Time): last update timestamp, server-assigned
- StatusChangedAt (time.Time): when the project last entered its current status, server-assigned

#### Scenario: Create project with required fields
- **WHEN** creating a Project with title "Launch website" and outcome "Website is live and accepting traffic"
- **THEN** system creates a Project with the specified title and outcome
- **AND** ID, CreatedAt, and UpdatedAt are server-assigned
- **AND** Status defaults to active
- **AND** StatusChangedAt equals CreatedAt

#### Scenario: Project title cannot be empty
- **WHEN** creating a Project with an empty title
- **THEN** system rejects the operation with a validation error

### Requirement: Project status values
A Project's Status SHALL be one of:
- active: project is in progress
- someday: project is parked for later consideration
- done: project is completed successfully
- dropped: project is abandoned

#### Scenario: Valid project statuses
- **WHEN** creating a Project with status active, someday, done, or dropped
- **THEN** system accepts the status value

#### Scenario: Invalid project status rejected
- **WHEN** creating a Project with an invalid status value
- **THEN** system rejects the operation with a validation error

### Requirement: Project status-change timestamp
A Project SHALL record StatusChangedAt, the instant it last entered its current status. On creation it SHALL equal CreatedAt (the transition into active). Every status transition (CompleteProject, DropProject, ParkProject, ReopenProject) SHALL overwrite StatusChangedAt with the instant supplied to that transition. UpdateProject SHALL NOT change StatusChangedAt (it does not change status). This mirrors Task.StatusChangedAt.

#### Scenario: Status change updates StatusChangedAt
- **WHEN** a transition (e.g. CompleteProject) is called with instant T
- **THEN** the project's StatusChangedAt is set to T

#### Scenario: Non-status update preserves StatusChangedAt
- **WHEN** UpdateProject is called
- **THEN** StatusChangedAt is unchanged

### Requirement: Project value semantics
Project SHALL use value semantics following domain conventions: no pointers in service interfaces, int64 for ID, nullable fields use pointer types, timestamps stored as UTC.

#### Scenario: Service returns Project value
- **WHEN** CreateProject is called
- **THEN** service returns Project value, not *Project

#### Scenario: Optional due date is nullable
- **WHEN** creating a Project without a due date
- **THEN** Project.Due is nil
