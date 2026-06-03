# project-list-presentation Specification

## Purpose
Defines how individual project rows are rendered in the project list, including status markers, title styling, task progress chips, due date chips, and selection highlighting.

## Requirements

### Requirement: Project row displays status marker
The system SHALL render each project row with a leading status marker indicating the project's current status.

#### Scenario: Open project marker
- **WHEN** a project has status "open"
- **THEN** the row SHALL display "[ ]" as the status marker

#### Scenario: Done project marker
- **WHEN** a project has status "done"
- **THEN** the row SHALL display "[x]" as the status marker

#### Scenario: Dropped project marker
- **WHEN** a project has status "dropped"
- **THEN** the row SHALL display "[-]" as the status marker

#### Scenario: Someday project marker
- **WHEN** a project has status "someday"
- **THEN** the row SHALL display "[?]" as the status marker

### Requirement: Project row displays title with status styling
The system SHALL render the project title after the status marker, styled according to project status.

#### Scenario: Open project title
- **WHEN** a project has status "open"
- **THEN** the title SHALL render in the default style

#### Scenario: Someday project title
- **WHEN** a project has status "someday"
- **THEN** the title SHALL render in a dimmed style

#### Scenario: Done project title
- **WHEN** a project has status "done"
- **THEN** the title SHALL render in a faint muted style

#### Scenario: Dropped project title
- **WHEN** a project has status "dropped"
- **THEN** the title SHALL render in a faint strikethrough style

### Requirement: Project row displays task progress chip
The system SHALL display a task progress chip showing the count of pending tasks versus non-dropped tasks linked to the project. Dropped tasks SHALL be excluded from both counts.

#### Scenario: Project with tasks
- **WHEN** a project has 3 pending tasks, 2 done tasks, and 1 dropped task
- **THEN** the row SHALL display a chip "3/5 tasks" (dropped excluded from total)

#### Scenario: Project with no non-dropped tasks
- **WHEN** a project has 0 non-dropped tasks
- **THEN** the row SHALL NOT display a task progress chip

#### Scenario: Project with all tasks complete
- **WHEN** a project has 0 pending tasks and 4 done tasks
- **THEN** the row SHALL display a chip "0/4 tasks"

### Requirement: Open project with no pending tasks shows warning chip
The system SHALL visually highlight an open project that has zero pending tasks, since this indicates a GTD process error (either a next action should be added, or the project status should change).

#### Scenario: Open project needs next action
- **WHEN** a project has status "open"
- **AND** the project has 0 pending tasks
- **THEN** the row SHALL display a "needs action" chip styled with the ready/teal urgency color

#### Scenario: Non-open project suppresses warning
- **WHEN** a project has status "someday", "done", or "dropped"
- **AND** the project has 0 pending tasks
- **THEN** no "needs action" chip SHALL be rendered

### Requirement: Project row displays due chip
The system SHALL display a due date chip with urgency coloring, following the same urgency palette as task due chips.

#### Scenario: Project with due date
- **WHEN** a project has a due date set
- **THEN** the row SHALL display a "due:<relative-time>" chip with urgency coloring

#### Scenario: Project overdue
- **WHEN** a project's due date has passed
- **AND** the project status is "open"
- **THEN** the row SHALL display an "overdue:<relative-time>" chip in red

#### Scenario: Project with no due date
- **WHEN** a project has no due date
- **THEN** no due chip SHALL be rendered

#### Scenario: Non-open project suppresses due chip
- **WHEN** a project has status "done" or "dropped"
- **THEN** the due chip SHALL NOT be rendered

### Requirement: Selected row is visually highlighted
The system SHALL indicate the currently selected project row with a cursor prefix and bold title.

#### Scenario: Selected row rendering
- **WHEN** a project row is the current selection
- **THEN** the row SHALL be prefixed with "> " and the title SHALL be bold

#### Scenario: Unselected row rendering
- **WHEN** a project row is not the current selection
- **THEN** the row SHALL be prefixed with "  "

### Requirement: Title truncation preserves chips
The system SHALL truncate the project title to fit within available width, keeping chips intact.

#### Scenario: Long title with chips
- **WHEN** the project title plus chips would exceed the row width
- **THEN** the title SHALL be truncated with "…" and chips SHALL remain fully visible