## MODIFIED Requirements

### Requirement: InboxService for clarify orchestration
The system SHALL provide an InboxService in the service/ package that orchestrates clarify operations across multiple stores. InboxService SHALL take store interfaces as dependencies.

#### Scenario: InboxService accepts store dependencies
- **WHEN** creating an InboxService
- **THEN** it accepts ItemStore, TaskStore, ProjectStore, ReferenceStore

#### Scenario: InboxService in service package
- **WHEN** InboxService is implemented
- **THEN** it resides in the service/ package

### Requirement: Item ClarifiedInto mutual exclusion
The Item entity SHALL have at most one ClarifiedInto field set at any time. A CHECK constraint SHALL enforce that at most one of `ClarifiedIntoTaskID`, `ClarifiedIntoProjectID`, `ClarifiedIntoReferenceID` is non-null, and Discarded is false when any ClarifiedInto is set. Incubate and ClarifyAsProject both target `ClarifiedIntoProjectID`; the project's status distinguishes the two outcomes.

#### Scenario: At most one ClarifiedInto
- **WHEN** an Item has ClarifiedIntoTaskID set
- **THEN** all other ClarifiedInto fields are null

#### Scenario: Discarded exclusive of ClarifiedInto
- **WHEN** an Item has Discarded=true
- **THEN** all ClarifiedInto fields are null

#### Scenario: CHECK constraint enforces exclusion
- **WHEN** attempting to set multiple ClarifiedInto fields in SQL
- **THEN** database rejects with CHECK constraint violation

## ADDED Requirements

### Requirement: FileAsReference clarify operation
The FileAsReference operation SHALL create a Reference entity for non-actionable content to keep for retrieval. The Item's `ClarifiedIntoReferenceID` SHALL point to the new Reference. The operation SHALL be transactional.

#### Scenario: FileAsReference creates reference
- **WHEN** FileAsReference is called with item ID and Reference data
- **THEN** the system creates a Reference entity
- **AND** Item.ClarifiedIntoReferenceID points to the new Reference

#### Scenario: FileAsReference is atomic
- **WHEN** FileAsReference is called
- **THEN** Reference creation and Item update occur in one transaction

#### Scenario: FileAsReference returns both entities
- **WHEN** FileAsReference is called successfully
- **THEN** it returns both the created Reference and updated Item

#### Scenario: FileAsReference copies item title by default
- **WHEN** FileAsReference is called without explicit title
- **THEN** the Reference title is copied from the Item title

#### Scenario: FileAsReference copies item description to body
- **WHEN** FileAsReference is called without explicit body
- **THEN** the Reference body is copied from the Item description

#### Scenario: FileAsReference fails for already-clarified item
- **WHEN** FileAsReference is called on an Item with ClarifiedInto already set
- **THEN** the system returns an error

#### Scenario: FileAsReference fails for discarded item
- **WHEN** FileAsReference is called on a discarded Item
- **THEN** the system returns an error
