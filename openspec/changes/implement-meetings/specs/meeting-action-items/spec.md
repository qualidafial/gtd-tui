## ADDED Requirements

### Requirement: MeetingService.AddActionItem
The system SHALL provide an AddActionItem method on MeetingService that atomically:
1. Creates an inbox Item with the provided title and optional description
2. Creates a MeetingLink connecting the Meeting to the new Item
3. Appends a uniform line to the Meeting body

All three operations SHALL occur in a single transaction.

#### Scenario: Add action item creates inbox item
- **WHEN** AddActionItem is called with a title
- **THEN** the system creates an inbox Item with that title
- **AND** the Item has nil ClarifiedInto pointers (still in inbox)

#### Scenario: Add action item creates MeetingLink
- **WHEN** AddActionItem is called
- **THEN** the system creates a MeetingLink with ItemID set
- **AND** the MeetingLink references the Meeting

#### Scenario: Add action item appends to meeting body
- **WHEN** AddActionItem is called with title "Follow up with Bob"
- **THEN** the system appends "- [ ] Follow up with Bob" to the Meeting body

#### Scenario: Add action item is atomic
- **WHEN** AddActionItem fails after creating the Item but before updating Meeting body
- **THEN** all changes are rolled back
- **AND** no orphaned Item or MeetingLink exists

#### Scenario: Add action item with description
- **WHEN** AddActionItem is called with title and description
- **THEN** the created Item has both title and description populated

### Requirement: AddActionItem return value
The AddActionItem method SHALL return the created Item and the updated Meeting.

#### Scenario: AddActionItem returns Item
- **WHEN** AddActionItem succeeds
- **THEN** the returned Item has ID, CreatedAt, UpdatedAt populated

#### Scenario: AddActionItem returns updated Meeting
- **WHEN** AddActionItem succeeds
- **THEN** the returned Meeting has UpdatedAt refreshed
- **AND** the Meeting body includes the appended action item line

### Requirement: AddActionItem validation
AddActionItem SHALL validate inputs before creating entities.

#### Scenario: AddActionItem rejects empty title
- **WHEN** AddActionItem is called with an empty title
- **THEN** the system rejects with a validation error

#### Scenario: AddActionItem rejects invalid meeting ID
- **WHEN** AddActionItem is called with a non-existent Meeting ID
- **THEN** the system returns a not found error

### Requirement: Action item line format
The uniform line appended to Meeting body SHALL use markdown checkbox format: `- [ ] <title>`. The line SHALL be appended on its own line, adding a newline before if the body does not end with a newline.

#### Scenario: Append to empty body
- **WHEN** AddActionItem is called on a Meeting with empty body
- **THEN** the body becomes "- [ ] <title>"

#### Scenario: Append to body ending with newline
- **WHEN** AddActionItem is called on a Meeting with body "Notes:\n"
- **THEN** the body becomes "Notes:\n- [ ] <title>"

#### Scenario: Append to body not ending with newline
- **WHEN** AddActionItem is called on a Meeting with body "Notes:"
- **THEN** the body becomes "Notes:\n- [ ] <title>"
