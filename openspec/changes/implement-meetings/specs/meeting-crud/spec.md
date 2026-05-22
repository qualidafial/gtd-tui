## ADDED Requirements

### Requirement: Meeting entity structure
The system SHALL provide a Meeting entity with the following fields:
- ID (int64, server-assigned)
- Title (string, non-empty)
- Body (string, markdown content)
- StartTime (time.Time, required)
- EndTime (time.Time, required)
- Attendees ([]string, stored as JSON)
- CreatedAt (time.Time, server-assigned)
- UpdatedAt (time.Time, server-assigned)

#### Scenario: Meeting has required fields
- **WHEN** a Meeting is created
- **THEN** the Meeting SHALL have ID, Title, StartTime, EndTime, CreatedAt, and UpdatedAt populated

#### Scenario: Meeting body is markdown
- **WHEN** a Meeting body contains markdown syntax
- **THEN** the system SHALL store the body as-is without modification

#### Scenario: Meeting attendees are string array
- **WHEN** a Meeting has attendees
- **THEN** attendees SHALL be stored as a JSON array of strings

### Requirement: MeetingService.CreateMeeting
The system SHALL provide a CreateMeeting method that creates a new Meeting. The method SHALL return the created Meeting with server-assigned fields (ID, CreatedAt, UpdatedAt) populated.

#### Scenario: Create meeting with valid data
- **WHEN** CreateMeeting is called with title, start time, end time, and optional body/attendees
- **THEN** the system creates the Meeting
- **AND** returns the Meeting with ID, CreatedAt, UpdatedAt populated

#### Scenario: Create meeting rejects empty title
- **WHEN** CreateMeeting is called with an empty title
- **THEN** the system rejects the request with a constraint error

#### Scenario: Create meeting rejects missing times
- **WHEN** CreateMeeting is called without StartTime or EndTime
- **THEN** the system rejects the request with a validation error

### Requirement: MeetingService.Meeting
The system SHALL provide a Meeting method that retrieves a single Meeting by ID.

#### Scenario: Get existing meeting
- **WHEN** Meeting is called with a valid ID
- **THEN** the system returns the Meeting

#### Scenario: Get non-existent meeting
- **WHEN** Meeting is called with an invalid ID
- **THEN** the system returns a not found error

### Requirement: MeetingService.Meetings
The system SHALL provide a Meetings method that retrieves a list of Meetings matching a filter. The filter SHALL support filtering by MeetingIDs.

#### Scenario: List all meetings
- **WHEN** Meetings is called with an empty filter
- **THEN** the system returns all Meetings

#### Scenario: List meetings by IDs
- **WHEN** Meetings is called with specific MeetingIDs
- **THEN** the system returns only Meetings with those IDs

### Requirement: MeetingService.UpdateMeeting
The system SHALL provide an UpdateMeeting method that updates an existing Meeting. The method SHALL return the updated Meeting with UpdatedAt refreshed.

#### Scenario: Update meeting title
- **WHEN** UpdateMeeting is called with a new title
- **THEN** the system updates the title
- **AND** returns the Meeting with UpdatedAt refreshed

#### Scenario: Update meeting body
- **WHEN** UpdateMeeting is called with a new body
- **THEN** the system updates the body

#### Scenario: Update meeting attendees
- **WHEN** UpdateMeeting is called with new attendees
- **THEN** the system updates the attendees JSON array

#### Scenario: Update non-existent meeting
- **WHEN** UpdateMeeting is called with an invalid ID
- **THEN** the system returns a not found error

### Requirement: MeetingService.DeleteMeeting
The system SHALL provide a DeleteMeeting method that deletes an existing Meeting and all associated MeetingLinks.

#### Scenario: Delete existing meeting
- **WHEN** DeleteMeeting is called with a valid ID
- **THEN** the system deletes the Meeting
- **AND** the system deletes all MeetingLinks referencing the Meeting

#### Scenario: Delete non-existent meeting
- **WHEN** DeleteMeeting is called with an invalid ID
- **THEN** the system returns a not found error

### Requirement: MeetingLink entity structure
The system SHALL provide a MeetingLink entity as a join row connecting a Meeting to exactly one of: Task, Project, or Item. The entity SHALL have:
- ID (int64, server-assigned)
- MeetingID (int64, required)
- TaskID (*int64, nullable)
- ProjectID (*int64, nullable)
- ItemID (*int64, nullable)
- CreatedAt (time.Time, server-assigned)

Exactly one of TaskID, ProjectID, or ItemID SHALL be set (enforced by CHECK constraint).

#### Scenario: MeetingLink to Task
- **WHEN** a MeetingLink is created with TaskID set
- **THEN** ProjectID and ItemID SHALL be null

#### Scenario: MeetingLink to Project
- **WHEN** a MeetingLink is created with ProjectID set
- **THEN** TaskID and ItemID SHALL be null

#### Scenario: MeetingLink to Item
- **WHEN** a MeetingLink is created with ItemID set
- **THEN** TaskID and ProjectID SHALL be null

#### Scenario: MeetingLink rejects multiple FKs
- **WHEN** a MeetingLink is created with multiple FKs set
- **THEN** the system rejects with a CHECK constraint violation

#### Scenario: MeetingLink rejects no FKs
- **WHEN** a MeetingLink is created with no FKs set
- **THEN** the system rejects with a CHECK constraint violation

### Requirement: MeetingLinks method
The system SHALL provide a MeetingLinks method that retrieves MeetingLinks for a given Meeting.

#### Scenario: Get meeting links
- **WHEN** MeetingLinks is called with a Meeting ID
- **THEN** the system returns all MeetingLinks for that Meeting
