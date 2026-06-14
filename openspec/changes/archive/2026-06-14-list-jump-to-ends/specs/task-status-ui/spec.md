## MODIFIED Requirements

### Requirement: Reorder limited to open tasks
The move bindings (shift+up / shift+down, shift+home / shift+end) SHALL be available only when the selected task is open, and only in the direction that keeps the task within the contiguous block of open tasks (which sort above closed ones). Move up and move-to-top SHALL be disabled when the selected task is the first task. Move down and move-to-bottom SHALL be disabled when the selected task is the last open task (the next task is closed, or there is none). When a binding is disabled it SHALL NOT be advertised in the help bar and pressing it SHALL have no effect.

#### Scenario: Move bindings hidden for a closed task
- **WHEN** the selected task is done or dropped
- **THEN** the help bar does not show the move-up/move-down or move-to-top/move-to-bottom bindings
- **AND** pressing shift+up, shift+down, shift+home, or shift+end does not reorder the task

#### Scenario: Move up disabled on the first task
- **WHEN** the selected open task is the first task in the list
- **THEN** the help bar does not show the move-up or move-to-top binding
- **AND** pressing shift+up or shift+home does not reorder the task

#### Scenario: Move down disabled on the last open task
- **WHEN** the selected open task is the last open task (the next task is closed or absent)
- **THEN** the help bar does not show the move-down or move-to-bottom binding
- **AND** pressing shift+down or shift+end does not reorder the task

#### Scenario: Move bindings available for an open task in the interior
- **WHEN** the selected task is open and has open tasks both above and below it
- **THEN** the help bar shows the move-up, move-down, move-to-top, and move-to-bottom bindings
- **AND** pressing them reorders the task

### Requirement: Reorder respects active task filter
On the task list, shift+up, shift+down, shift+home, and shift+end SHALL pass the list's currently active TaskFilter into MoveTaskUp / MoveTaskDown / MoveTaskFirst / MoveTaskLast. A shift+up/shift+down press SHALL produce a one-slot move within the visible (filtered) list; a shift+home/shift+end press SHALL move the task to the first/last position of the visible (filtered) list. Items outside the filter are not consulted when picking the new position and MAY interleave with filtered items as a result.

#### Scenario: Move down within a search filter takes one press to visibly shift
- **WHEN** the query filter narrows the task list to a subset of open tasks
- **AND** user presses shift+down on a non-last filtered open task
- **THEN** the task SHALL appear one slot lower in the visible list after the reload

#### Scenario: move last within a search filter jumps to the last filtered slot
- **WHEN** the query filter narrows the task list to a subset of open tasks
- **AND** user presses shift+end on a non-last filtered open task
- **THEN** the task SHALL appear after every other filtered open task in the visible list after the reload
- **AND** the cursor SHALL stay on the moved task

#### Scenario: move first within a search filter jumps to the first filtered slot
- **WHEN** the query filter narrows the task list to a subset of open tasks
- **AND** user presses shift+home on a non-first filtered open task
- **THEN** the task SHALL appear before every other filtered open task in the visible list after the reload
- **AND** the cursor SHALL stay on the moved task

#### Scenario: In-project task list move stays within the project
- **WHEN** the task list is rendered inside a project view (using the projectTaskService wrapper)
- **AND** user presses shift+home on a non-first open task
- **THEN** the move SHALL reorder only within that project's open tasks even if the list's TaskFilter has no ProjectID set
