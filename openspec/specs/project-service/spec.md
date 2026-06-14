# project-service Specification

## Purpose
Defines the `ProjectService` interface and its operations: create/update/get/list, status transitions (complete, drop, park, reopen) with atomicity, and project reordering.

## Requirements

### Requirement: ProjectService interface
The system SHALL provide a ProjectService interface with CRUD operations, status transitions, and reordering. The interface SHALL be defined in the root package alongside the Project domain type.

#### Scenario: ProjectService in root package
- **WHEN** implementing project management
- **THEN** ProjectService interface is defined in the gtd root package

### Requirement: Create project operation
ProjectService SHALL provide CreateProject(ctx, Project) (Project, error) that creates a new project. The returned Project SHALL have ID, CreatedAt, and UpdatedAt populated. Status SHALL default to open if not specified. Open and someday projects SHALL be assigned a fractional order key appended after all existing projects of the same status.

#### Scenario: Create project returns populated value
- **WHEN** CreateProject is called with a valid Project
- **THEN** the returned Project has ID, CreatedAt, UpdatedAt populated
- **AND** Status is open if not specified

#### Scenario: Create project persists data
- **WHEN** CreateProject is called
- **THEN** the project is persisted to the database

#### Scenario: Create open project assigns order key
- **WHEN** CreateProject is called with open status
- **THEN** the project is assigned an order key after all existing open projects

#### Scenario: Create someday project assigns order key
- **WHEN** CreateProject is called with someday status
- **THEN** the project is assigned an order key after all existing someday projects

### Requirement: Update project operation
ProjectService SHALL provide UpdateProject(ctx, Project) (Project, error) that updates an existing project. The returned Project SHALL have UpdatedAt refreshed.

#### Scenario: Update project refreshes timestamp
- **WHEN** UpdateProject is called
- **THEN** the returned Project has UpdatedAt refreshed

### Requirement: Get project operation
ProjectService SHALL provide GetProject(ctx, id int64) (Project, error) that retrieves a project by ID. If the project does not exist, an error SHALL be returned.

#### Scenario: Get existing project
- **WHEN** GetProject is called with a valid ID
- **THEN** the Project is returned

#### Scenario: Get non-existent project
- **WHEN** GetProject is called with an invalid ID
- **THEN** an error is returned indicating project not found

### Requirement: List projects operation
ProjectService SHALL provide ListProjects(ctx, filter ProjectFilter) ([]Project, error) that returns projects matching the filter criteria. The filter SHALL support filtering by status. Results SHALL be ordered in three tiers: open projects first (by order_key ASC), then someday projects (by order_key ASC), then done/dropped projects (by status_changed_at DESC). Within each tier, ties are broken by id ASC.

#### Scenario: List all projects
- **WHEN** ListProjects is called with no filter
- **THEN** all projects are returned: open first by order key, then someday by order key, then done/dropped by status_changed_at descending

#### Scenario: List projects by status
- **WHEN** ListProjects is called with status filter open
- **THEN** only open projects are returned

### Requirement: Complete project transition
ProjectService SHALL provide CompleteProject(ctx, id int64, cascade bool, at time.Time) (Project, error) that transitions a project to done status, clears its order key, and records the supplied `at` instant as the project's StatusChangedAt. The cascade flag determines handling of pending tasks. When cascade marks tasks done, the same `at` instant SHALL be recorded as each task's StatusChangedAt.

#### Scenario: Complete project with cascade
- **WHEN** CompleteProject is called with cascade=true
- **THEN** Project status is set to done
- **AND** Project order key is cleared
- **AND** all pending tasks under the project are marked done
- **AND** each cascaded task's StatusChangedAt is set to the supplied instant

#### Scenario: Complete project with detach
- **WHEN** CompleteProject is called with cascade=false
- **THEN** Project status is set to done
- **AND** Project order key is cleared
- **AND** all pending tasks have ProjectID set to nil
- **AND** tasks become standalone

### Requirement: Drop project transition
ProjectService SHALL provide DropProject(ctx, id int64, cascade bool, at time.Time) (Project, error) that transitions a project to dropped status, clears its order key, and records the supplied `at` instant as the project's StatusChangedAt. The cascade flag determines handling of pending tasks. When cascade marks tasks dropped, the same `at` instant SHALL be recorded as each task's StatusChangedAt.

#### Scenario: Drop project with cascade
- **WHEN** DropProject is called with cascade=true
- **THEN** Project status is set to dropped
- **AND** Project order key is cleared
- **AND** all pending tasks under the project are marked dropped
- **AND** each cascaded task's StatusChangedAt is set to the supplied instant

#### Scenario: Drop project with detach
- **WHEN** DropProject is called with cascade=false
- **THEN** Project status is set to dropped
- **AND** Project order key is cleared
- **AND** all pending tasks have ProjectID set to nil
- **AND** tasks become standalone

### Requirement: Park project transition
ProjectService SHALL provide ParkProject(ctx, id int64, at time.Time) (Project, error) that transitions a project to someday status, assigns a fresh order key appended after all existing someday projects, and records the supplied instant as the project's StatusChangedAt. Task statuses SHALL NOT change; only default view filtering is affected.

#### Scenario: Park project
- **WHEN** ParkProject is called on an open project
- **THEN** Project status is set to someday
- **AND** a fresh order key is assigned within the someday ordering
- **AND** task statuses remain unchanged

### Requirement: Reopen project transition
ProjectService SHALL provide ReopenProject(ctx, id int64, at time.Time) (Project, error) that transitions a project from a non-open status (someday, done, or dropped) back to open, assigns a fresh order key appended after all existing open projects, and records the supplied instant as the project's StatusChangedAt. Reopen mirrors ReopenTask: it restores the project to open and SHALL NOT change the status of any tasks under the project.

#### Scenario: Reopen parked project
- **WHEN** ReopenProject is called on a someday project
- **THEN** Project status is set to open
- **AND** a fresh order key is assigned

#### Scenario: Reopen completed project
- **WHEN** ReopenProject is called on a done project
- **THEN** Project status is set to open
- **AND** a fresh order key is assigned

#### Scenario: Reopen dropped project
- **WHEN** ReopenProject is called on a dropped project
- **THEN** Project status is set to open
- **AND** a fresh order key is assigned

#### Scenario: Reopen does not change task statuses
- **WHEN** ReopenProject is called on a project with done or dropped tasks
- **THEN** all task statuses remain unchanged

### Requirement: Transition atomicity
All status transition methods SHALL be transactional. The project status change, order key update, and task cascade/detach SHALL occur atomically.

#### Scenario: Transition rollback on failure
- **WHEN** CompleteProject fails during task cascade
- **THEN** the project status change is rolled back
- **AND** no tasks are modified

### Requirement: Project reordering
ProjectService SHALL provide `MoveProjectUp(ctx context.Context, id int64, filter ProjectFilter) error` and `MoveProjectDown(ctx context.Context, id int64, filter ProjectFilter) error` to shift a project one position within projects of the same status that also match `filter`, and `MoveProjectFirst(ctx context.Context, id int64, filter ProjectFilter) error` and `MoveProjectLast(ctx context.Context, id int64, filter ProjectFilter) error` to move a project ahead of / after every same-status project that matches `filter`. The moving project's status group is always the universe — open projects are reordered among open projects; someday projects are reordered among someday projects. The supplied `filter` (Search, and any Status that matches the moving project's status) SHALL narrow the candidate neighbors further. Reordering uses fractional-indexed order keys (same `orderkey` package as tasks) with a renumber fallback when keys are exhausted; on exhaustion the renumber SHALL act on the entire same-status group (not just the filtered subset), preserving every non-moving project's relative position. All four moves SHALL be rejected for done/dropped projects. The new position SHALL be computed against the *filtered* set so a move is relative to the visible list; same-status projects outside the filter MAY interleave with filtered projects as a result, and on key exhaustion the moving project may visibly jump several positions in unfiltered views. `MoveProjectFirst` on the first filtered project and `MoveProjectLast` on the last filtered project SHALL be no-ops.

#### Scenario: Move open project up
- **WHEN** MoveProjectUp is called on an open project that is not first among open projects matching the filter
- **THEN** the project moves one position earlier among the filtered open projects

#### Scenario: Move someday project down
- **WHEN** MoveProjectDown is called on a someday project that is not last among someday projects matching the filter
- **THEN** the project moves one position later among the filtered someday projects

#### Scenario: Move open project first
- **WHEN** MoveProjectFirst is called on an open project that is not already first among open projects matching the filter
- **THEN** the project SHALL receive an order_key earlier than every other filtered open project
- **AND** open projects that do not match the filter SHALL retain their existing order_keys

#### Scenario: Move someday project last
- **WHEN** MoveProjectLast is called on a someday project that is not already last among someday projects matching the filter
- **THEN** the project SHALL receive an order_key later than every other filtered someday project
- **AND** projects that do not match the filter SHALL retain their existing order_keys

#### Scenario: Move to boundary is a no-op
- **WHEN** MoveProjectFirst is called on the first filtered project of its status group, or MoveProjectLast on the last
- **THEN** no order_keys SHALL change

#### Scenario: Reorder rejected for done/dropped project
- **WHEN** MoveProjectUp, MoveProjectDown, MoveProjectFirst, or MoveProjectLast is called on a done or dropped project
- **THEN** an error is returned

#### Scenario: Move down within a search filter
- **WHEN** MoveProjectDown is called on an open project with a Search filter that matches a subset of open projects
- **THEN** the project SHALL receive a new order_key between the next filtered project and the one after it
- **AND** open projects that do not match the filter SHALL retain their existing order_keys

#### Scenario: Key exhaustion renumbers the entire same-status group
- **WHEN** any project move is called and `orderkey.Between` cannot produce a key strictly between the filtered prev/next neighbors
- **THEN** every project in the moving project's status group SHALL be assigned a fresh evenly-spaced order_key in its current order, with the moving project slotted at its target position relative to its filtered neighbors
- **AND** the relative order of every non-moving project in that status group SHALL be preserved
