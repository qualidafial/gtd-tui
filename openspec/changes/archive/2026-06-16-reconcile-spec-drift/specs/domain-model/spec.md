## MODIFIED Requirements

### Requirement: Task entity for actionable items
The system SHALL provide a Task entity representing a single actionable item. A Task SHALL have:
- Status: `open`, `done`, or `dropped`
- Optional Assignee (a person's name); a non-nil Assignee marks the task as **delegated** (waiting on someone else), while a nil Assignee is a plain next action. There is no separate `Kind` field — delegation is inferred from Assignee.
- Optional Due date (firm deadline)
- Optional DeferUntil date (soft "don't show until" date)
- Optional ProjectID (0..1 relationship to Project)

#### Scenario: Create next action task
- **WHEN** user creates a task with no assignee
- **THEN** system creates a Task with open status and a nil Assignee

#### Scenario: Create delegated task
- **WHEN** user creates a task with an assignee
- **THEN** system creates a Task with open status and a non-nil Assignee, marking it delegated

#### Scenario: Deferred tasks filter from default views
- **WHEN** a Task has DeferUntil in the future
- **THEN** the Task SHALL be filtered out of default task views
- **AND** the Task status remains open
