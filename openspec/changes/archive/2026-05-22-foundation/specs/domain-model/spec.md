## ADDED Requirements

### Requirement: Item entity for inbox capture
The system SHALL provide an Item entity representing unprocessed captures in the inbox. An Item SHALL have title, description, and timestamps. When clarified, an Item SHALL be soft-deleted via a ClarifiedInto pointer to whatever it became (Task, Project, Someday, Reference). Discarded items SHALL be marked rather than hard-deleted.

#### Scenario: Create inbox item
- **WHEN** user captures a thought to the inbox
- **THEN** system creates an Item with title, optional description, and timestamps

#### Scenario: Clarify item preserves lineage
- **WHEN** an Item is clarified into a Task
- **THEN** the Item's ClarifiedInto pointer references the new Task
- **AND** the Item remains queryable for timeline history

### Requirement: Task entity for actionable items
The system SHALL provide a Task entity representing a single actionable item. A Task SHALL have:
- Kind: `next_action` (do ASAP) or `delegated` (waiting on someone else, with Assignee string)
- Status: `pending`, `done`, or `dropped`
- Optional Due date (firm deadline)
- Optional DeferUntil date (soft "don't show until" date)
- Optional ProjectID (0..1 relationship to Project)

#### Scenario: Create next action task
- **WHEN** user creates a task with kind next_action
- **THEN** system creates a Task with pending status and no assignee

#### Scenario: Create delegated task
- **WHEN** user creates a task with kind delegated
- **THEN** system creates a Task with pending status and an Assignee string

#### Scenario: Deferred tasks filter from default views
- **WHEN** a Task has DeferUntil in the future
- **THEN** the Task SHALL be filtered out of default task views
- **AND** the Task status remains pending

### Requirement: Project entity for multi-step outcomes
The system SHALL provide a Project entity representing a multi-step outcome. A Project SHALL have:
- Title (short, for lists)
- Outcome statement (desired end state)
- Description
- Optional Due date
- Status: `active`, `someday` (parked), `done`, or `dropped`

#### Scenario: Create active project
- **WHEN** user creates a project
- **THEN** system creates a Project with active status

#### Scenario: Complete project cascades to tasks
- **WHEN** user completes a project with cascade flag
- **THEN** all pending tasks under the project are marked done

#### Scenario: Complete project detaches tasks
- **WHEN** user completes a project with detach flag
- **THEN** all pending tasks have ProjectID set to nil
- **AND** tasks become standalone

#### Scenario: Park project filters tasks
- **WHEN** user parks a project (status someday)
- **THEN** tasks under the project are filtered from default views
- **AND** task statuses remain unchanged

#### Scenario: No pending tasks under closed projects
- **WHEN** a project transitions to done or dropped
- **THEN** no pending tasks SHALL remain under the project

### Requirement: Someday entity for parked ideas
The system SHALL provide a Someday entity representing a parked idea not yet fleshed out enough to be a Project. A Someday SHALL have title, description, and ReviewedAt timestamp (defaults to creation time, used to surface stalest items in periodic review). A Someday SHALL be promotable to Project or Task.

#### Scenario: Create someday item
- **WHEN** user incubates an idea
- **THEN** system creates a Someday with ReviewedAt set to creation time

#### Scenario: Promote someday to project
- **WHEN** user promotes a Someday to Project
- **THEN** system creates a Project from the Someday

### Requirement: Reference entity for retrieval content
The system SHALL provide a Reference entity representing standalone markdown content kept for retrieval (recipes, config snippets, link dumps). A Reference SHALL have title and body. References are NOT linked to projects or tasks.

#### Scenario: File as reference
- **WHEN** user files an inbox item as reference
- **THEN** system creates a Reference with title and body

### Requirement: Meeting entity for meeting records
The system SHALL provide a Meeting entity for meeting records. A Meeting SHALL have:
- Title
- Body (markdown discussion notes)
- Required start/end times
- Attendees (JSON string array)

A Meeting cross-references projects and tasks via MeetingLink. Action items captured during a meeting spawn inbox Items with MeetingLink references.

#### Scenario: Create meeting record
- **WHEN** user creates a meeting
- **THEN** system creates a Meeting with title, times, and attendees

#### Scenario: Add action item to meeting
- **WHEN** user adds an action item during a meeting
- **THEN** system creates an inbox Item linked to the Meeting via MeetingLink
- **AND** system appends a uniform line to the Meeting body

### Requirement: Comment entity for contextual notes
The system SHALL provide a Comment entity for short, event-shaped text attached to exactly one Task or Project. Comments SHALL be spawned implicitly by edits (the comment parameter on UpdateTask/UpdateProject) and explicitly via the comment API. Comments SHALL be editable.

#### Scenario: Edit with comment
- **WHEN** user edits a task with a comment
- **THEN** system creates a Comment attached to the Task
- **AND** the edit and comment are recorded atomically

#### Scenario: Standalone comment
- **WHEN** user adds a comment without editing
- **THEN** system creates a Comment for context that isn't tied to a metadata change

### Requirement: MeetingLink for meeting cross-references
The system SHALL provide a MeetingLink entity as a join row connecting a Meeting to a Task, Project, or Item. Exactly one target FK SHALL be set (enforced by CHECK). When an Item is clarified, its MeetingLink SHALL be rewritten to point at the resulting entity.

#### Scenario: Meeting links to project
- **WHEN** a meeting is linked to a project
- **THEN** MeetingLink has ProjectID set and other FKs null

#### Scenario: MeetingLink follows clarification
- **WHEN** an Item with a MeetingLink is clarified into a Task
- **THEN** the MeetingLink is rewritten to point at the Task

### Requirement: Task to Project relationship
A Task SHALL belong to zero or one Projects via ProjectID. Standalone tasks have nil ProjectID. The reverse relationship (Project → Tasks) is derived from Task.ProjectID.

#### Scenario: Task belongs to project
- **WHEN** a task is assigned to a project
- **THEN** Task.ProjectID references the project

#### Scenario: Standalone task
- **WHEN** a task has no project
- **THEN** Task.ProjectID is nil

### Requirement: Value semantics for domain types
All domain types SHALL use value semantics (no pointers in service interfaces). IDs SHALL be int64. Nullable timestamps and FKs SHALL use pointer types. Timestamps SHALL be stored as UTC.

#### Scenario: Service returns value
- **WHEN** CreateTask is called
- **THEN** service returns Task value, not *Task
