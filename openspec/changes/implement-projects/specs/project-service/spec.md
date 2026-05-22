## ADDED Requirements

### Requirement: ProjectService interface
The system SHALL provide a ProjectService interface with CRUD operations and status transitions. The interface SHALL be defined in the root package alongside the Project domain type.

#### Scenario: ProjectService in root package
- **WHEN** implementing project management
- **THEN** ProjectService interface is defined in the gtd root package

### Requirement: Create project operation
ProjectService SHALL provide CreateProject(ctx, Project) (Project, error) that creates a new project. The returned Project SHALL have ID, CreatedAt, and UpdatedAt populated. Status SHALL default to active if not specified.

#### Scenario: Create project returns populated value
- **WHEN** CreateProject is called with a valid Project
- **THEN** the returned Project has ID, CreatedAt, UpdatedAt populated
- **AND** Status is active if not specified

#### Scenario: Create project persists data
- **WHEN** CreateProject is called
- **THEN** the project is persisted to the database

### Requirement: Update project operation
ProjectService SHALL provide UpdateProject(ctx, Project, comment string) (Project, error) that updates an existing project. The returned Project SHALL have UpdatedAt refreshed. If comment is non-empty, a Comment entity SHALL be created atomically.

#### Scenario: Update project refreshes timestamp
- **WHEN** UpdateProject is called
- **THEN** the returned Project has UpdatedAt refreshed

#### Scenario: Update project with comment
- **WHEN** UpdateProject is called with a non-empty comment
- **THEN** a Comment is created attached to the Project
- **AND** the update and comment creation are atomic

#### Scenario: Update project without comment
- **WHEN** UpdateProject is called with an empty comment
- **THEN** no Comment is created

### Requirement: Get project operation
ProjectService SHALL provide GetProject(ctx, id int64) (Project, error) that retrieves a project by ID. If the project does not exist, an error SHALL be returned.

#### Scenario: Get existing project
- **WHEN** GetProject is called with a valid ID
- **THEN** the Project is returned

#### Scenario: Get non-existent project
- **WHEN** GetProject is called with an invalid ID
- **THEN** an error is returned indicating project not found

### Requirement: List projects operation
ProjectService SHALL provide ListProjects(ctx, filter ProjectFilter) ([]Project, error) that returns projects matching the filter criteria. The filter SHALL support filtering by status.

#### Scenario: List all projects
- **WHEN** ListProjects is called with no filter
- **THEN** all projects are returned

#### Scenario: List projects by status
- **WHEN** ListProjects is called with status filter active
- **THEN** only active projects are returned

### Requirement: Complete project transition
ProjectService SHALL provide CompleteProject(ctx, id int64, cascade bool, comment string) (Project, error) that transitions a project to done status. The cascade flag determines handling of pending tasks.

#### Scenario: Complete project with cascade
- **WHEN** CompleteProject is called with cascade=true
- **THEN** Project status is set to done
- **AND** all pending tasks under the project are marked done

#### Scenario: Complete project with detach
- **WHEN** CompleteProject is called with cascade=false
- **THEN** Project status is set to done
- **AND** all pending tasks have ProjectID set to nil
- **AND** tasks become standalone

#### Scenario: Complete project with comment
- **WHEN** CompleteProject is called with a non-empty comment
- **THEN** a Comment is created attached to the Project

### Requirement: Drop project transition
ProjectService SHALL provide DropProject(ctx, id int64, cascade bool, comment string) (Project, error) that transitions a project to dropped status. The cascade flag determines handling of pending tasks.

#### Scenario: Drop project with cascade
- **WHEN** DropProject is called with cascade=true
- **THEN** Project status is set to dropped
- **AND** all pending tasks under the project are marked dropped

#### Scenario: Drop project with detach
- **WHEN** DropProject is called with cascade=false
- **THEN** Project status is set to dropped
- **AND** all pending tasks have ProjectID set to nil
- **AND** tasks become standalone

#### Scenario: Drop project with comment
- **WHEN** DropProject is called with a non-empty comment
- **THEN** a Comment is created attached to the Project

### Requirement: Park project transition
ProjectService SHALL provide ParkProject(ctx, id int64, comment string) (Project, error) that transitions a project to someday status. Task statuses SHALL NOT change; only default view filtering is affected.

#### Scenario: Park project
- **WHEN** ParkProject is called on an active project
- **THEN** Project status is set to someday
- **AND** task statuses remain unchanged

#### Scenario: Park project with comment
- **WHEN** ParkProject is called with a non-empty comment
- **THEN** a Comment is created attached to the Project

### Requirement: Unpark project transition
ProjectService SHALL provide UnparkProject(ctx, id int64, comment string) (Project, error) that transitions a project from someday back to active status.

#### Scenario: Unpark project
- **WHEN** UnparkProject is called on a someday project
- **THEN** Project status is set to active

#### Scenario: Unpark project with comment
- **WHEN** UnparkProject is called with a non-empty comment
- **THEN** a Comment is created attached to the Project

### Requirement: Transition atomicity
All status transition methods SHALL be transactional. The project status change, task cascade/detach, and optional comment creation SHALL all occur atomically.

#### Scenario: Transition rollback on failure
- **WHEN** CompleteProject fails during task cascade
- **THEN** the project status change is rolled back
- **AND** no tasks are modified
