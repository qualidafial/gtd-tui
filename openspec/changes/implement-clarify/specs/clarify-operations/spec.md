## ADDED Requirements

### Requirement: InboxService for clarify orchestration
The system SHALL provide an InboxService in the service/ package that orchestrates clarify operations across multiple stores. InboxService SHALL take store interfaces as dependencies.

#### Scenario: InboxService accepts store dependencies
- **WHEN** creating an InboxService
- **THEN** it accepts ItemStore, TaskStore, ProjectStore, SomedayStore, ReferenceStore

#### Scenario: InboxService in service package
- **WHEN** InboxService is implemented
- **THEN** it resides in the service/ package

### Requirement: Discard clarify operation
The Discard operation SHALL mark an Item as discarded for non-actionable, unwanted captures. No destination entity is created. The operation SHALL be transactional.

#### Scenario: Discard marks item
- **WHEN** Discard is called with an item ID
- **THEN** the Item's Discarded field is set to true
- **AND** no destination entity is created

#### Scenario: Discard is transactional
- **WHEN** Discard is called
- **THEN** the Item update occurs in a single transaction

#### Scenario: Discard returns updated item
- **WHEN** Discard is called successfully
- **THEN** it returns the updated Item with Discarded=true

#### Scenario: Discard fails for already-clarified item
- **WHEN** Discard is called on an Item with ClarifiedInto already set
- **THEN** the system returns an error

#### Scenario: Discard fails for non-existent item
- **WHEN** Discard is called with non-existent item ID
- **THEN** the system returns an error

### Requirement: Incubate clarify operation
The Incubate operation SHALL create a Someday entity for non-actionable items to revisit later. The Item's ClarifiedIntoSomedayID SHALL point to the new Someday. The operation SHALL be transactional.

#### Scenario: Incubate creates someday
- **WHEN** Incubate is called with item ID and Someday data
- **THEN** the system creates a Someday entity
- **AND** Item.ClarifiedIntoSomedayID points to the new Someday

#### Scenario: Incubate is atomic
- **WHEN** Incubate is called
- **THEN** Someday creation and Item update occur in one transaction

#### Scenario: Incubate returns both entities
- **WHEN** Incubate is called successfully
- **THEN** it returns both the created Someday and updated Item

#### Scenario: Incubate copies item title by default
- **WHEN** Incubate is called without explicit title
- **THEN** the Someday title is copied from the Item title

#### Scenario: Incubate fails for already-clarified item
- **WHEN** Incubate is called on an Item with ClarifiedInto already set
- **THEN** the system returns an error

#### Scenario: Incubate rollback on failure
- **WHEN** Incubate fails during Someday creation
- **THEN** no changes are persisted

### Requirement: FileAsReference clarify operation
The FileAsReference operation SHALL create a Reference entity for non-actionable content to keep for retrieval. The Item's ClarifiedIntoReferenceID SHALL point to the new Reference. The operation SHALL be transactional.

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

### Requirement: ClarifyAsTask clarify operation
The ClarifyAsTask operation SHALL create a Task entity for actionable single-step items. The task kind and optional project are specified at clarify time. The Item's ClarifiedIntoTaskID SHALL point to the new Task. The operation SHALL be transactional.

#### Scenario: ClarifyAsTask creates next action
- **WHEN** ClarifyAsTask is called with kind=next_action
- **THEN** the system creates a Task with kind next_action
- **AND** Item.ClarifiedIntoTaskID points to the new Task

#### Scenario: ClarifyAsTask creates delegated task
- **WHEN** ClarifyAsTask is called with kind=delegated and Assignee
- **THEN** the system creates a Task with kind delegated and the Assignee
- **AND** Item.ClarifiedIntoTaskID points to the new Task

#### Scenario: ClarifyAsTask with project
- **WHEN** ClarifyAsTask is called with a ProjectID
- **THEN** the created Task has ProjectID set

#### Scenario: ClarifyAsTask is atomic
- **WHEN** ClarifyAsTask is called
- **THEN** Task creation and Item update occur in one transaction

#### Scenario: ClarifyAsTask returns both entities
- **WHEN** ClarifyAsTask is called successfully
- **THEN** it returns both the created Task and updated Item

#### Scenario: ClarifyAsTask copies item title by default
- **WHEN** ClarifyAsTask is called without explicit title
- **THEN** the Task title is copied from the Item title

#### Scenario: ClarifyAsTask do-it-now creates done task
- **WHEN** ClarifyAsTask is called with Status=done
- **THEN** the created Task has Status done
- **AND** timeline preserves the captured-to-done transition

#### Scenario: ClarifyAsTask fails for already-clarified item
- **WHEN** ClarifyAsTask is called on an Item with ClarifiedInto already set
- **THEN** the system returns an error

#### Scenario: ClarifyAsTask fails for invalid project
- **WHEN** ClarifyAsTask is called with non-existent ProjectID
- **THEN** the system returns an error

### Requirement: ClarifyAsProject clarify operation
The ClarifyAsProject operation SHALL create a Project entity for actionable multi-step outcomes. The Item's ClarifiedIntoProjectID SHALL point to the new Project. The operation SHALL be transactional.

#### Scenario: ClarifyAsProject creates project
- **WHEN** ClarifyAsProject is called with item ID and Project data
- **THEN** the system creates a Project with status active
- **AND** Item.ClarifiedIntoProjectID points to the new Project

#### Scenario: ClarifyAsProject is atomic
- **WHEN** ClarifyAsProject is called
- **THEN** Project creation and Item update occur in one transaction

#### Scenario: ClarifyAsProject returns both entities
- **WHEN** ClarifyAsProject is called successfully
- **THEN** it returns both the created Project and updated Item

#### Scenario: ClarifyAsProject copies item title by default
- **WHEN** ClarifyAsProject is called without explicit title
- **THEN** the Project title is copied from the Item title

#### Scenario: ClarifyAsProject fails for already-clarified item
- **WHEN** ClarifyAsProject is called on an Item with ClarifiedInto already set
- **THEN** the system returns an error

### Requirement: Item ClarifiedInto mutual exclusion
The Item entity SHALL have at most one ClarifiedInto field set at any time. A CHECK constraint SHALL enforce that at most one of ClarifiedIntoTaskID, ClarifiedIntoProjectID, ClarifiedIntoSomedayID, ClarifiedIntoReferenceID is non-null, and Discarded is false when any ClarifiedInto is set.

#### Scenario: At most one ClarifiedInto
- **WHEN** an Item has ClarifiedIntoTaskID set
- **THEN** all other ClarifiedInto fields are null

#### Scenario: Discarded exclusive of ClarifiedInto
- **WHEN** an Item has Discarded=true
- **THEN** all ClarifiedInto fields are null

#### Scenario: CHECK constraint enforces exclusion
- **WHEN** attempting to set multiple ClarifiedInto fields in SQL
- **THEN** database rejects with CHECK constraint violation

### Requirement: Clarify operations prevent double-clarification
All clarify operations SHALL fail if the Item has already been clarified (any ClarifiedInto field set or Discarded=true).

#### Scenario: Cannot clarify discarded item
- **WHEN** any clarify operation is called on a discarded Item
- **THEN** the system returns an error

#### Scenario: Cannot re-clarify item
- **WHEN** any clarify operation is called on an already-clarified Item
- **THEN** the system returns an error
