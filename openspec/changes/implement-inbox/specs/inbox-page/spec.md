## ADDED Requirements

### Requirement: Inbox page for unprocessed items
The system SHALL provide an inbox screen at `tui/pages/inbox/` that displays unclarified, non-discarded inbox items. The screen SHALL be a Screen (per `tui-application`) registered as a tab in the root tabContainer alongside Tasks and Projects.

#### Scenario: Display inbox items
- **WHEN** the inbox screen is active
- **THEN** it displays Items returned by InboxService.List in FIFO order (oldest first)

#### Scenario: Empty inbox state
- **WHEN** InboxService.List returns no items
- **THEN** the inbox screen renders an empty-state message rather than an empty list

#### Scenario: Select item for inspection
- **WHEN** the user moves the cursor onto an item
- **THEN** the item's title and description are visible in the active row

### Requirement: Inbox tab registration
The system SHALL register an "Inbox" tab in the root tabContainer so the inbox screen is reachable via tab navigation.

#### Scenario: Inbox tab present
- **WHEN** the application starts
- **THEN** the tabContainer SHALL include an "Inbox" tab in addition to "Tasks" and "Projects"
