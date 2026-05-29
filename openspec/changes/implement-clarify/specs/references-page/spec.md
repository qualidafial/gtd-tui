## ADDED Requirements

### Requirement: References page for retrieval content
The system SHALL provide a references screen at `tui/pages/references/` that displays Reference entities and supports filtering by title via the shared querybar component.

#### Scenario: Display references
- **WHEN** the references screen is active
- **THEN** it displays Reference entries returned by ReferenceStore.List

#### Scenario: Filter references by title
- **WHEN** the user types a query in the querybar
- **THEN** the displayed list narrows to references whose title matches the query

### Requirement: References tab registration
The system SHALL register a "References" tab in the root tabContainer.

#### Scenario: References tab present
- **WHEN** the application starts
- **THEN** the tabContainer SHALL include a "References" tab
