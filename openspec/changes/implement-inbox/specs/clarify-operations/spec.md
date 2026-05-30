## ADDED Requirements

### Requirement: InboxService for clarify orchestration
The system SHALL provide an InboxService in the service/ package that orchestrates clarify operations across multiple stores. InboxService SHALL take store interfaces as dependencies.

#### Scenario: InboxService accepts store dependencies
- **WHEN** creating an InboxService
- **THEN** it accepts ItemStore, TaskStore, ProjectStore (ReferenceStore is added by `implement-references`)

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
The Incubate operation SHALL create a Project with `Status=someday` for ideas to revisit later. The Item's `ClarifiedIntoProjectID` SHALL point to the new Project. The operation SHALL be transactional. Someday items are not a separate entity — they reuse the Project entity per the finalized `project-entity` spec.

#### Scenario: Incubate creates someday project
- **WHEN** Incubate is called with item ID and project data
- **THEN** the system creates a Project with `Status=someday`
- **AND** Item.ClarifiedIntoProjectID points to the new Project

#### Scenario: Incubate is atomic
- **WHEN** Incubate is called
- **THEN** Project creation and Item update occur in one transaction

#### Scenario: Incubate returns both entities
- **WHEN** Incubate is called successfully
- **THEN** it returns both the created Project and updated Item

#### Scenario: Incubate copies item title by default
- **WHEN** Incubate is called without explicit title
- **THEN** the Project title is copied from the Item title

#### Scenario: Incubate fails for already-clarified item
- **WHEN** Incubate is called on an Item with ClarifiedInto already set
- **THEN** the system returns an error

#### Scenario: Incubate rollback on failure
- **WHEN** Incubate fails during Project creation
- **THEN** no changes are persisted

#### Scenario: Incubated project later reopened
- **WHEN** the user invokes ReopenProject on a project that was created via Incubate
- **THEN** the project transitions from someday to open per the existing project-service semantics
- **AND** no further action is required on the originating Item

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
The ClarifyAsProject operation SHALL create a Project with `Status=open` for actionable multi-step outcomes. The Item's `ClarifiedIntoProjectID` SHALL point to the new Project. The operation SHALL be transactional. This differs from Incubate only in the project's initial status.

#### Scenario: ClarifyAsProject creates project
- **WHEN** ClarifyAsProject is called with item ID and Project data
- **THEN** the system creates a Project with `Status=open`
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
The Item entity SHALL have at most one ClarifiedInto field set at any time. A CHECK constraint SHALL enforce that at most one of `ClarifiedIntoTaskID`, `ClarifiedIntoProjectID` is non-null, and Discarded is false when either is set. Incubate and ClarifyAsProject both target `ClarifiedIntoProjectID`; the project's status distinguishes the two outcomes. `implement-references` extends this requirement to include `ClarifiedIntoReferenceID`.

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
