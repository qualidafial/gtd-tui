# project-list-ui Specification

## Purpose
Defines the interactive behavior of the project list screen, including loading, navigation, keybindings, and project state transitions available from the list.

## Requirements

### Requirement: Project list loads projects on init
The system SHALL load all projects from ProjectService on initialization using the current query filter and display them in the list sorted by status group (open, someday, then done/dropped) and order_key within each group.

#### Scenario: Initial load
- **WHEN** the project list screen initializes
- **THEN** it SHALL issue a command to load all projects via ListProjects with the current parsed filter
- **AND** display the results in the list

#### Scenario: Reload on init after overlay dismiss
- **WHEN** an overlay is dismissed and the project list receives Init
- **THEN** it SHALL reload the project list from the service using the current filter

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
Pressing esc while editing SHALL revert the query bar's text to the last applied query, blur the bar, and cause the project list to reload using that previously-applied query. The list SHALL end in the same state as immediately after the last commit, undoing any live-previewed filter introduced by debounced typing since.

#### Scenario: Cancel reverts text and snaps list back to last applied query
- **WHEN** the user edits the query and presses esc
- **THEN** the query bar reverts to the last applied query
- **AND** the project list reflects that previously-applied query
- **AND** the query bar is no longer focused

### Requirement: Help bar reflects editing state
While the query bar is focused, the help bar SHALL show the editing actions (apply on Enter, cancel on Esc) instead of the list navigation bindings.

#### Scenario: Help bar while editing
- **WHEN** the query bar is focused
- **THEN** the help bar shows the apply (Enter) and cancel (Esc) bindings

#### Scenario: Help bar while not editing
- **WHEN** the query bar is not focused
- **THEN** the help bar shows the list navigation and project-action bindings

### Requirement: Project list loads task counts
The system SHALL load task counts (pending and total) for all displayed projects after the project list loads.

#### Scenario: Task counts loaded
- **WHEN** projects are loaded
- **THEN** the system SHALL fetch task counts for all loaded project IDs in a single batch query
- **AND** associate counts with each project row for rendering

### Requirement: Quick-create project with "+" or "insert" key
The system SHALL allow creating a new project by pressing "+" or "insert", which pushes the project edit overlay with an empty project. On submit, it creates an open project with the entered fields.

#### Scenario: Create new project
- **WHEN** user presses "+" or "insert"
- **THEN** the project edit overlay SHALL be pushed with an empty project
- **WHEN** user fills in fields and submits
- **THEN** the system SHALL call CreateProject with status=open and the entered fields
- **AND** the editor overlay SHALL dismiss
- **AND** the project view screen SHALL be pushed for the newly created project

#### Scenario: Cancel project creation
- **WHEN** user presses "+" or "insert" to open the editor
- **AND** user presses escape
- **THEN** the overlay SHALL be dismissed without creating a project

### Requirement: Edit project from project list with "e" key
The system SHALL allow editing the selected project by pressing "e", which pushes the project edit overlay with the selected project. The "e" key SHALL be enabled when a project is selected and disabled when the list is empty.

#### Scenario: Edit selected project
- **WHEN** user selects a project
- **AND** presses "e"
- **THEN** the project edit overlay SHALL be pushed with the selected project

#### Scenario: Edit disabled when no selection
- **WHEN** the list is empty
- **THEN** the "e" keybinding SHALL be disabled

#### Scenario: Return from edit
- **WHEN** user saves or cancels the edit
- **THEN** the overlay SHALL dismiss
- **AND** the project list SHALL reinitialize

### Requirement: Enter project view with enter key
The system SHALL allow entering a project's detail view by pressing enter on the selected project, which pushes the project view screen as an overlay. The keymap field carrying this binding SHALL be named `View` (the key remains `enter`).

#### Scenario: Enter project view
- **WHEN** user selects a project
- **AND** presses enter
- **THEN** the project view screen SHALL be pushed as an overlay
- **AND** the project view SHALL display the selected project's details and tasks

#### Scenario: Return from project view
- **WHEN** user presses esc in the project view
- **THEN** the project view SHALL dismiss
- **AND** the project list SHALL reinitialize

#### Scenario: Binding field name reflects action, not key
- **WHEN** code or tests reference the view binding via the projects keymap
- **THEN** the field SHALL be `projects.KeyMap.View`, not `Enter`

### Requirement: Toggle project status with space key
The system SHALL allow toggling project status with space: completing open projects (with confirmation), and reopening someday/done/dropped projects (immediately). The keymap field carrying this binding SHALL be named `ToggleComplete` (the key is `space`; the displayed label flips between `complete` and `reopen` via `SetHelp`).

#### Scenario: Complete open project
- **WHEN** user selects an open project
- **AND** presses space
- **THEN** a confirmation overlay SHALL be pushed
- **WHEN** user confirms
- **THEN** CompleteProject SHALL be called with cascade=true
- **AND** the overlay SHALL dismiss

#### Scenario: Reopen someday project with space
- **WHEN** user selects a someday project
- **AND** presses space
- **THEN** ReopenProject SHALL be called immediately (no confirmation)
- **AND** the list SHALL reload

#### Scenario: Reopen done/dropped project with space
- **WHEN** user selects a done or dropped project
- **AND** presses space
- **THEN** ReopenProject SHALL be called immediately (no confirmation)
- **AND** the list SHALL reload

#### Scenario: Binding field name reflects primary action
- **WHEN** code or tests reference the toggle binding via the projects keymap
- **THEN** the field SHALL be `projects.KeyMap.ToggleComplete`, not `Toggle`

### Requirement: Drop project with delete key
The system SHALL allow dropping an open or someday project by pressing delete/backspace, which pushes a confirmation overlay.

#### Scenario: Drop project
- **WHEN** user selects an open or someday project
- **AND** presses delete
- **THEN** a confirmation overlay SHALL be pushed showing task cascade info
- **WHEN** user confirms
- **THEN** DropProject SHALL be called with cascade=true
- **AND** the overlay SHALL dismiss

#### Scenario: Drop disabled for done/dropped
- **WHEN** user selects a done or dropped project
- **THEN** the delete keybinding SHALL be disabled

### Requirement: Park project with "s" key
The system SHALL allow parking an open project by pressing "s" (someday), transitioning it immediately without confirmation.

#### Scenario: Park open project
- **WHEN** user selects an open project
- **AND** presses "s"
- **THEN** ParkProject SHALL be called
- **AND** the list SHALL reload

#### Scenario: Park disabled for non-open
- **WHEN** user selects a someday, done, or dropped project
- **THEN** the "s" keybinding SHALL be disabled

### Requirement: Reorder projects with shift+up/down
The system SHALL allow reordering open and someday projects within their status group using shift+up and shift+down (one slot) and shift+home and shift+end (to the first / last position of the status group). The reorder SHALL be scoped to the list's currently active project filter: the call to MoveProjectUp / MoveProjectDown / MoveProjectFirst / MoveProjectLast SHALL pass the list's current ProjectFilter so a shift+up/down press shifts the project one slot, and a shift+home/end press moves it to the first/last position, within the visible (filtered) list rather than within the full status group.

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

### Requirement: Keybindings reflect selected project state
The system SHALL enable/disable action keybindings based on the selected project's status, and update help text accordingly.

#### Scenario: Open project selected
- **WHEN** an open project is selected
- **THEN** space label SHALL be "complete", delete SHALL be enabled, "s" SHALL be enabled, reorder SHALL be conditionally enabled

#### Scenario: Someday project selected
- **WHEN** a someday project is selected
- **THEN** space label SHALL be "reopen", delete SHALL be enabled, "s" SHALL be disabled, reorder SHALL be conditionally enabled

#### Scenario: Done project selected
- **WHEN** a done project is selected
- **THEN** space label SHALL be "reopen", delete SHALL be disabled, "s" SHALL be disabled, reorder SHALL be disabled

#### Scenario: No project selected
- **WHEN** the list is empty
- **THEN** all action bindings except "+" (new) SHALL be disabled

### Requirement: Revert to default project filter with backslash
The project list SHALL reset the query bar to the default project query
`status:open` when the user presses backslash (`\`) while the query bar is not
focused (the list is focused), reparsing it via `projectquery.Parse` and
reloading the list with the resulting ProjectFilter.
The reset SHALL record the default as the applied query, so that a subsequent Esc
reverts to the default rather than to the previously-typed query. The `\` binding
SHALL be disabled — hidden from the help bar and inert when pressed — whenever the
current query already equals the default query.

#### Scenario: Revert restores the default project query
- **WHEN** the user has applied a different query and the query bar is not focused
- **AND** the user presses `\`
- **THEN** the query bar text SHALL revert to `status:open`
- **AND** the list SHALL reload showing only open projects
- **AND** the applied query SHALL become `status:open`

#### Scenario: Esc after revert reverts to the default
- **WHEN** the user has pressed `\` to revert to the default
- **AND** the user later focuses the query bar, edits it, and presses Esc
- **THEN** the query bar SHALL revert to `status:open`

#### Scenario: Binding disabled when already at default
- **WHEN** the current query equals `status:open`
- **THEN** the `\` binding SHALL be disabled and SHALL NOT appear in the help bar

#### Scenario: Binding inert while editing
- **WHEN** the query bar is focused for editing
- **AND** the user types `\`
- **THEN** the filter SHALL NOT be reset
- **AND** the `\` character SHALL be entered into the query