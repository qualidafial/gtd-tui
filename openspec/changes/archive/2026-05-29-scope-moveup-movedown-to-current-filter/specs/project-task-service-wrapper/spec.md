## ADDED Requirements

### Requirement: Wrapper injects ProjectID on move
The projectTaskService wrapper's MoveTaskUp and MoveTaskDown SHALL inject the wrapper's project ID into the supplied filter's ProjectID field before delegating to the inner TaskService. This keeps the wrapper's "scoped to one project" invariant consistent with ListTasks and CreateTask: a move requested in an in-project view always reorders within that project even if the caller passed an unfiltered or differently-scoped TaskFilter.

#### Scenario: MoveTaskUp scoped to project
- **WHEN** MoveTaskUp is called on the wrapper with a TaskFilter whose ProjectID is nil
- **THEN** the filter SHALL have ProjectID set to the wrapper's project ID before delegation

#### Scenario: MoveTaskDown overrides foreign ProjectID
- **WHEN** MoveTaskDown is called on the wrapper with a TaskFilter whose ProjectID points to a different project
- **THEN** the filter's ProjectID SHALL be overwritten with the wrapper's project ID before delegation

## MODIFIED Requirements

### Requirement: Wrapper delegates all other methods
All other TaskService methods (GetTask, UpdateTask, CompleteTask, DropTask, ReopenTask, DeleteTask) SHALL delegate to the inner service without modification. MoveTaskUp and MoveTaskDown are NOT in this set — they inject the wrapper's ProjectID into the filter (see "Wrapper injects ProjectID on move").

#### Scenario: UpdateTask delegates unchanged
- **WHEN** UpdateTask is called on the wrapper
- **THEN** the call SHALL be delegated to the inner TaskService without modifying the task

#### Scenario: CompleteTask delegates unchanged
- **WHEN** CompleteTask is called on the wrapper
- **THEN** the call SHALL be delegated to the inner TaskService without modification
