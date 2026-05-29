## ADDED Requirements

### Requirement: Meetings page for meeting records
The system SHALL provide a meetings screen at `tui/pages/meetings/` that displays Meeting entities. By default the screen SHALL surface upcoming and recent meetings (within a configurable window centered on the current time).

#### Scenario: Display meetings
- **WHEN** the meetings screen is active
- **THEN** it displays Meeting entries returned by MeetingService.List

#### Scenario: Default surfaces relevant meetings
- **WHEN** the meetings screen loads with no filter
- **THEN** upcoming meetings and recent past meetings are shown, ordered by start time

### Requirement: Meetings tab registration
The system SHALL register a "Meetings" tab in the root tabContainer.

#### Scenario: Meetings tab present
- **WHEN** the application starts
- **THEN** the tabContainer SHALL include a "Meetings" tab
