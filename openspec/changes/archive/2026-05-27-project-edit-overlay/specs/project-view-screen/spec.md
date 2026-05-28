## ADDED Requirements

### Requirement: Edit project from project view
Pressing `e` in the project view (while the task list is not capturing input) SHALL push the project edit overlay for the current project.

#### Scenario: Open project editor
- **WHEN** the user presses `e` in the project view
- **AND** the task list is not capturing input
- **THEN** the project edit overlay SHALL be pushed with the current project

#### Scenario: Edit key ignored when capturing input
- **WHEN** the user presses `e` while the task list is capturing input (e.g., query bar active)
- **THEN** the keypress SHALL pass through to the task list

### Requirement: Reload project header after edit
When the project edit overlay dismisses, the project view SHALL re-fetch the project from the service and update its header to reflect any changes to title, outcome, or due.

#### Scenario: Header updates after edit
- **WHEN** the user edits the project title and saves
- **AND** the editor overlay dismisses
- **THEN** the project view header SHALL display the updated title

#### Scenario: Header updates after outcome change
- **WHEN** the user edits the project outcome and saves
- **AND** the editor overlay dismisses
- **THEN** the project view header SHALL display the updated outcome