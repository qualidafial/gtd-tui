## MODIFIED Requirements

### Requirement: Reorder projects with shift+up/down
The system SHALL allow reordering open and someday projects within their status group using shift+up and shift+down (one slot) and shift+home and shift+end (to the top / bottom of the status group). The reorder SHALL be scoped to the list's currently active project filter: the call to MoveProjectUp / MoveProjectDown / MoveProjectFirst / MoveProjectLast SHALL pass the list's current ProjectFilter so a shift+up/down press shifts the project one slot, and a shift+home/end press moves it to the first/last position, within the visible (filtered) list rather than within the full status group.

#### Scenario: Move project up
- **WHEN** user selects an open project that is not the first open project in the visible list
- **AND** presses shift+up
- **THEN** MoveProjectUp SHALL be called with the list's current ProjectFilter
- **AND** the list SHALL reload with the cursor on the moved project

#### Scenario: Move project down
- **WHEN** user selects an open project that is not the last open project in the visible list
- **AND** presses shift+down
- **THEN** MoveProjectDown SHALL be called with the list's current ProjectFilter
- **AND** the list SHALL reload with the cursor on the moved project

#### Scenario: Move project first
- **WHEN** user selects an open project that is not the first open project in the visible list
- **AND** presses shift+home
- **THEN** MoveProjectFirst SHALL be called with the list's current ProjectFilter
- **AND** the list SHALL reload with the cursor on the moved project, now first among the filtered projects of its status group

#### Scenario: Move project last
- **WHEN** user selects an open project that is not the last open project in the visible list
- **AND** presses shift+end
- **THEN** MoveProjectLast SHALL be called with the list's current ProjectFilter
- **AND** the list SHALL reload with the cursor on the moved project, now last among the filtered projects of its status group

#### Scenario: Move disabled at boundary
- **WHEN** user selects the first project of its status group in the visible list
- **THEN** shift+up and shift+home SHALL be disabled
- **WHEN** user selects the last project of its status group in the visible list
- **THEN** shift+down and shift+end SHALL be disabled

#### Scenario: Move disabled for done/dropped
- **WHEN** user selects a done or dropped project
- **THEN** shift+up, shift+down, shift+home, and shift+end SHALL be disabled

#### Scenario: Move down within a search filter takes one press to visibly shift
- **WHEN** the query filter narrows the list to a subset of open projects
- **AND** user presses shift+down on a non-last filtered project
- **THEN** the project SHALL appear one slot lower in the visible list after the reload
