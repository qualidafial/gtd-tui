## ADDED Requirements

### Requirement: MeetingLink follows Item through clarification
When an Item with a MeetingLink is clarified into a Task or Project, the MeetingLink SHALL be rewritten to point at the resulting entity. This preserves the meeting's provenance trail permanently.

#### Scenario: ClarifyAsTask rewrites MeetingLink
- **WHEN** an Item with a MeetingLink is clarified into a Task
- **THEN** the MeetingLink is updated to set TaskID to the new Task ID
- **AND** the MeetingLink ItemID is set to null

#### Scenario: ClarifyAsProject rewrites MeetingLink
- **WHEN** an Item with a MeetingLink is clarified into a Project
- **THEN** the MeetingLink is updated to set ProjectID to the new Project ID
- **AND** the MeetingLink ItemID is set to null

### Requirement: MeetingLink rewriting is transactional
The MeetingLink rewriting SHALL occur within the same transaction as the clarify operation.

#### Scenario: Clarify transaction includes link rewrite
- **WHEN** ClarifyAsTask creates a Task and rewrites a MeetingLink
- **THEN** both operations occur in a single transaction

#### Scenario: Clarify rollback reverts link rewrite
- **WHEN** ClarifyAsTask fails after rewriting MeetingLink
- **THEN** the MeetingLink is rolled back to its original state

### Requirement: Non-linked clarification paths
Clarify operations that do not produce a Task or Project (Discard, Incubate, FileAsReference) SHALL leave MeetingLinks pointing at the Item.

#### Scenario: Discard preserves MeetingLink to Item
- **WHEN** an Item with a MeetingLink is discarded
- **THEN** the MeetingLink continues to point at the Item
- **AND** the Item is marked as discarded

#### Scenario: Incubate rewrites MeetingLink to Project
- **WHEN** an Item with a MeetingLink is incubated
- **THEN** a Project with `Status=someday` is created
- **AND** the MeetingLink is rewritten to point at the new Project (same target type as ClarifyAsProject; only the project status differs)

#### Scenario: FileAsReference preserves MeetingLink to Item
- **WHEN** an Item with a MeetingLink is filed as reference
- **THEN** the MeetingLink continues to point at the Item
- **AND** a Reference entity is created

### Requirement: Multiple MeetingLinks on one Item
An Item may have multiple MeetingLinks (from multiple meetings). When clarified, all MeetingLinks referencing the Item SHALL be rewritten.

#### Scenario: Clarify Item with multiple MeetingLinks
- **WHEN** an Item with MeetingLinks from two different Meetings is clarified into a Task
- **THEN** both MeetingLinks are updated to point at the Task

### Requirement: Item deletion cascades to MeetingLinks
When an Item is hard-deleted, associated MeetingLinks SHALL be deleted via ON DELETE CASCADE.

#### Scenario: Delete Item cascades to MeetingLink
- **WHEN** an Item with a MeetingLink is hard-deleted
- **THEN** the MeetingLink is also deleted
