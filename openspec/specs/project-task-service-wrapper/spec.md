# project-task-service-wrapper Specification

## Purpose
Defines the behavior of the projectTaskService wrapper, which scopes TaskService operations to a specific project by injecting the project's ID into list and create calls.

## Requirements

### Requirement: Wrapper injects ProjectID on list
The projectTaskService wrapper SHALL implement gtd.TaskService. ListTasks SHALL inject the project's ID into the filter's ProjectID field before delegating to the inner service.

#### Scenario: ListTasks scoped to project
- **WHEN** ListTasks is called on the wrapper
- **THEN** the filter SHALL have ProjectID set to the wrapper's project ID
- **AND** the call SHALL be delegated to the inner TaskService

### Requirement: Wrapper injects ProjectID on create
CreateTask SHALL set the task's ProjectID to the wrapper's project ID before delegating to the inner service.

#### Scenario: CreateTask stamps project
- **WHEN** CreateTask is called on the wrapper with a task that has nil ProjectID
- **THEN** the task SHALL have ProjectID set to the wrapper's project ID
- **AND** the call SHALL be delegated to the inner TaskService

### Requirement: Wrapper delegates all other methods
All other TaskService methods (GetTask, UpdateTask, CompleteTask, DropTask, ReopenTask, DeleteTask, MoveUp, MoveDown) SHALL delegate to the inner service without modification.

#### Scenario: UpdateTask delegates unchanged
- **WHEN** UpdateTask is called on the wrapper
- **THEN** the call SHALL be delegated to the inner TaskService without modifying the task

#### Scenario: CompleteTask delegates unchanged
- **WHEN** CompleteTask is called on the wrapper
- **THEN** the call SHALL be delegated to the inner TaskService without modification