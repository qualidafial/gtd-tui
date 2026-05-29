## MODIFIED Requirements

### Requirement: Project list loads projects on init
The system SHALL load all projects from ProjectService on initialization using the current query filter and display them in the list sorted by status group (open, someday, then done/dropped) and order_key within each group.

#### Scenario: Initial load
- **WHEN** the project list screen initializes
- **THEN** it SHALL issue a command to load all projects via ListProjects with the current parsed filter
- **AND** display the results in the list

#### Scenario: Reload on init after overlay dismiss
- **WHEN** an overlay is dismissed and the project list receives Init
- **THEN** it SHALL reload the project list from the service using the current filter

## ADDED Requirements

### Requirement: Query bar on the project list
The project list SHALL display a query bar above the list using the shared `querybar` component. The bar SHALL be seeded with `status:open` on startup. When the query is empty, the bar SHALL show a placeholder indicating that all projects are shown.

#### Scenario: Default query on startup
- **WHEN** the project list first loads
- **THEN** the query bar shows `status:open`
- **AND** only open projects are listed

#### Scenario: Query bar is single line
- **WHEN** the project list renders
- **THEN** the query bar SHALL occupy exactly one line above the list

### Requirement: Focus and edit the project query
Pressing `/` SHALL focus the query bar for editing. The list's built-in filter keybinding SHALL be disabled so the query bar is the only filtering mechanism.

#### Scenario: Focus query bar
- **WHEN** the user presses `/`
- **THEN** the query bar becomes focused and editable

### Requirement: Apply project query on enter
Pressing enter while the query bar is focused SHALL parse the query via `projectquery.Parse` and, on success, reload the list with the parsed `ProjectFilter`. On parse failure, the error SHALL be displayed in the app error bar.

#### Scenario: Apply a valid query
- **WHEN** the user types `status:someday` and presses enter
- **THEN** the list reloads showing only someday projects

#### Scenario: Apply with free-text search
- **WHEN** the user types `shed` and presses enter
- **THEN** the list reloads showing only projects matching "shed" in title, outcome, or description

#### Scenario: Enter on invalid query shows error
- **WHEN** the user presses enter on `status:bogus`
- **THEN** the list is not reloaded
- **AND** the error is displayed in the app error bar

### Requirement: Cancel project query edit on esc
Pressing esc while editing SHALL revert the query bar to the last applied query without reloading.

#### Scenario: Cancel reverts text
- **WHEN** the user edits the query and presses esc
- **THEN** the query bar reverts to the last applied query
- **AND** the listed projects are unchanged

### Requirement: Help bar reflects editing state
While the query bar is focused, the help bar SHALL show the editing actions (apply on Enter, cancel on Esc) instead of the list navigation bindings.

#### Scenario: Help bar while editing
- **WHEN** the query bar is focused
- **THEN** the help bar shows the apply (Enter) and cancel (Esc) bindings

#### Scenario: Help bar while not editing
- **WHEN** the query bar is not focused
- **THEN** the help bar shows the list navigation and project-action bindings
