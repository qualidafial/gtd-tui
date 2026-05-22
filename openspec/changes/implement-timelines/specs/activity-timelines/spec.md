## ADDED Requirements

### Requirement: TimelineEntry domain type
The system SHALL provide a TimelineEntry domain type representing a discrete event in the system. A TimelineEntry SHALL have:
- ID (int64)
- EntityType (string: "task", "project", "meeting", "item", "comment", "someday")
- EntityID (int64)
- EventType (string: "created", "updated", "status_changed", "clarified", "commented")
- Details (map[string]any for event-specific metadata)
- CreatedAt (time.Time)

TimelineEntry SHALL use value semantics consistent with other domain types.

#### Scenario: TimelineEntry has required fields
- **WHEN** a TimelineEntry is created
- **THEN** it has EntityType, EntityID, EventType, and CreatedAt populated

#### Scenario: TimelineEntry details vary by event type
- **WHEN** a status_changed event is recorded
- **THEN** Details contains "from" and "to" status values

### Requirement: TimelineService interface
The system SHALL provide a TimelineService interface for recording and querying timeline entries. The service SHALL be defined in the root package alongside other service interfaces.

#### Scenario: TimelineService in root package
- **WHEN** defining TimelineService
- **THEN** it is placed in the root gtd package

### Requirement: Record timeline entry
The TimelineService SHALL provide a Record method to create a timeline entry. The method SHALL accept a TimelineEntry and return the entry with server-assigned fields (ID, CreatedAt) populated.

#### Scenario: Record timeline entry
- **WHEN** Record is called with a TimelineEntry
- **THEN** the entry is persisted
- **AND** the returned entry has ID and CreatedAt populated

### Requirement: Query entity timeline
The TimelineService SHALL provide a method to query timeline entries for a specific entity. The method SHALL accept entity type and entity ID, returning entries in chronological order (oldest first).

#### Scenario: Query task timeline
- **WHEN** querying timeline for entity_type="task" and entity_id=5
- **THEN** system returns all timeline entries for that task
- **AND** entries are ordered by CreatedAt ascending

#### Scenario: Empty timeline returns empty slice
- **WHEN** querying timeline for an entity with no events
- **THEN** system returns an empty slice, not an error

### Requirement: Query global timeline
The TimelineService SHALL provide a method to query timeline entries across all entities. The method SHALL support pagination via limit and cursor, returning entries in reverse chronological order (newest first).

#### Scenario: Query global timeline
- **WHEN** querying global timeline with limit=50
- **THEN** system returns up to 50 most recent entries
- **AND** entries are ordered by CreatedAt descending

#### Scenario: Paginate global timeline
- **WHEN** querying global timeline with a cursor from previous results
- **THEN** system returns the next page of entries

### Requirement: Timeline entry for entity creation
The system SHALL generate a timeline entry with event_type "created" when a Task, Project, Meeting, Item, Someday, or Comment is created.

#### Scenario: Task creation generates timeline entry
- **WHEN** CreateTask is called
- **THEN** a timeline entry is created with event_type="created" and entity_type="task"

#### Scenario: Project creation generates timeline entry
- **WHEN** CreateProject is called
- **THEN** a timeline entry is created with event_type="created" and entity_type="project"

#### Scenario: Meeting creation generates timeline entry
- **WHEN** CreateMeeting is called
- **THEN** a timeline entry is created with event_type="created" and entity_type="meeting"

#### Scenario: Item creation generates timeline entry
- **WHEN** an inbox Item is captured
- **THEN** a timeline entry is created with event_type="created" and entity_type="item"

### Requirement: Timeline entry for entity updates
The system SHALL generate a timeline entry with event_type "updated" when a Task, Project, or Meeting is updated with non-status field changes.

#### Scenario: Task update generates timeline entry
- **WHEN** UpdateTask is called with changed fields
- **THEN** a timeline entry is created with event_type="updated"
- **AND** Details contains the changed field names

#### Scenario: Update with only status change uses status_changed
- **WHEN** a task status changes
- **THEN** event_type is "status_changed", not "updated"

### Requirement: Timeline entry for status transitions
The system SHALL generate a timeline entry with event_type "status_changed" when a Task or Project status changes. The Details SHALL include "from" and "to" values.

#### Scenario: Complete task generates status_changed entry
- **WHEN** CompleteTask is called
- **THEN** a timeline entry is created with event_type="status_changed"
- **AND** Details contains {from: "pending", to: "done"}

#### Scenario: Drop project generates status_changed entry
- **WHEN** DropProject is called
- **THEN** a timeline entry is created with event_type="status_changed"
- **AND** Details contains {from: <previous_status>, to: "dropped"}

#### Scenario: Park project generates status_changed entry
- **WHEN** ParkProject is called
- **THEN** a timeline entry is created with event_type="status_changed"
- **AND** Details contains {from: "active", to: "someday"}

### Requirement: Timeline entry for clarification
The system SHALL generate timeline entries when an inbox Item is clarified. Two entries SHALL be created:
1. On the Item: event_type "clarified" with Details containing the destination type and ID
2. On the destination entity: event_type "created" with Details containing the source item ID

#### Scenario: Clarify item as task generates two entries
- **WHEN** ClarifyAsTask is called for Item 5, creating Task 10
- **THEN** Item 5 gets entry with event_type="clarified", Details={into_type: "task", into_id: 10}
- **AND** Task 10 gets entry with event_type="created", Details={from_item: 5}

#### Scenario: Clarify item as project generates two entries
- **WHEN** ClarifyAsProject is called for Item 5, creating Project 3
- **THEN** Item 5 gets entry with event_type="clarified", Details={into_type: "project", into_id: 3}
- **AND** Project 3 gets entry with event_type="created", Details={from_item: 5}

#### Scenario: Discard item generates clarified entry
- **WHEN** Discard is called for an Item
- **THEN** Item gets entry with event_type="clarified", Details={discarded: true}

### Requirement: Timeline entry for comments
The system SHALL generate a timeline entry with event_type "commented" when a Comment is added to a Task or Project. The entry SHALL be on the parent entity (Task or Project), not the Comment itself, with Details containing the comment ID.

#### Scenario: Comment on task generates timeline entry
- **WHEN** a Comment is added to Task 5
- **THEN** Task 5 gets entry with event_type="commented", Details={comment_id: <id>}

#### Scenario: Edit-with-comment generates both entries
- **WHEN** UpdateTask is called with a comment
- **THEN** Task gets both an "updated" entry and a "commented" entry

### Requirement: Timeline entries are transactional
Timeline entry generation SHALL occur within the same database transaction as the operation that triggered it. If the operation fails, no timeline entry SHALL be persisted.

#### Scenario: Failed operation creates no timeline entry
- **WHEN** CreateTask fails due to validation error
- **THEN** no timeline entry is created

#### Scenario: Timeline entry in same transaction
- **WHEN** CompleteTask succeeds
- **THEN** the status change and timeline entry are committed atomically

### Requirement: Timeline entries are immutable
Timeline entries SHALL NOT be editable or deletable through the service interface. They serve as an immutable audit log.

#### Scenario: No update method for timeline entries
- **WHEN** inspecting TimelineService interface
- **THEN** there is no UpdateTimelineEntry method

#### Scenario: No delete method for timeline entries
- **WHEN** inspecting TimelineService interface
- **THEN** there is no DeleteTimelineEntry method

### Requirement: Project timeline includes task events
When querying a Project's timeline, the system SHALL include timeline entries for Tasks belonging to that Project. This provides a unified view of project activity.

#### Scenario: Project timeline shows task creation
- **WHEN** querying timeline for Project 3
- **AND** Task 10 belongs to Project 3
- **THEN** results include Task 10's timeline entries

#### Scenario: Project timeline shows task completion
- **WHEN** a task under Project 3 is completed
- **THEN** Project 3's timeline includes the status_changed entry

### Requirement: Meeting timeline includes linked entity events
When querying a Meeting's timeline, the system SHALL include timeline entries for entities linked via MeetingLink. This shows the meeting's action items and their progress.

#### Scenario: Meeting timeline shows action item clarification
- **WHEN** querying timeline for Meeting 2
- **AND** Item 5 was captured from Meeting 2 and clarified into Task 10
- **THEN** results include Item 5's clarified entry and Task 10's created entry

### Requirement: SQLite storage for timeline entries
The SQLite implementation SHALL store timeline entries in a `timeline_entries` table with appropriate indexes for efficient querying.

#### Scenario: Index on entity lookup
- **WHEN** querying by entity_type and entity_id
- **THEN** query uses an index, not a full table scan

#### Scenario: Index on created_at for global timeline
- **WHEN** querying global timeline with ORDER BY created_at
- **THEN** query uses an index for ordering
