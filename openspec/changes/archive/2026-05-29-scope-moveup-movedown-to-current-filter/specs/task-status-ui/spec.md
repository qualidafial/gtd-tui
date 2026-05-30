## ADDED Requirements

### Requirement: Reorder respects active task filter
On the task list, shift+up and shift+down SHALL pass the list's currently active TaskFilter into MoveTaskUp / MoveTaskDown. A single press SHALL produce a one-slot move within the visible (filtered) list rather than within all open tasks. Items outside the filter are not consulted when picking the new position and MAY interleave with filtered items as a result.

#### Scenario: Move down within a search filter takes one press to visibly shift
- **WHEN** the query filter narrows the task list to a subset of open tasks
- **AND** user presses shift+down on a non-last filtered open task
- **THEN** the task SHALL appear one slot lower in the visible list after the reload

#### Scenario: In-project task list move stays within the project
- **WHEN** the task list is rendered inside a project view (using the projectTaskService wrapper)
- **AND** user presses shift+up on a non-first open task
- **THEN** the move SHALL reorder only within that project's open tasks even if the list's TaskFilter has no ProjectID set
