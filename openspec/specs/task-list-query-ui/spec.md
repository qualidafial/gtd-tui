# task-list-query-ui Specification

## Purpose
TBD - created by syncing change task-query-filter. Update Purpose after archive.

## Requirements

### Requirement: Query bar on the task list
The task list SHALL display a query bar showing the active query string. The bar SHALL be seeded with the default query `status:pending ready:now` on startup. When the query is empty, the bar SHALL show a placeholder indicating that all tasks are shown.

#### Scenario: Default query on startup
- **WHEN** the task list first loads
- **THEN** the query bar shows `status:pending ready:now`
- **AND** only pending tasks that are available now (not deferred to the future) are listed

#### Scenario: Empty query shows placeholder
- **WHEN** the query bar is empty
- **THEN** a placeholder indicating all tasks are shown (e.g. `(all tasks)`) is displayed in the bar

### Requirement: Focus and edit the query
Pressing `/` SHALL focus the query bar for editing. The list's built-in filter keybinding SHALL be disabled so the query bar is the only filtering mechanism.

#### Scenario: Focus query bar
- **WHEN** the user presses `/`
- **THEN** the query bar becomes focused and editable

### Requirement: Help bar reflects editing state
While the query bar is focused for editing, the help bar SHALL show the editing actions (apply on Enter, cancel on Esc) instead of the list navigation bindings. When the query bar is not focused, the help bar SHALL show the list bindings.

#### Scenario: Help bar while editing
- **WHEN** the query bar is focused
- **THEN** the help bar shows the apply (Enter) and cancel (Esc) bindings

#### Scenario: Help bar while not editing
- **WHEN** the query bar is not focused
- **THEN** the help bar shows the list navigation and task-action bindings

### Requirement: Global keys suppressed while editing
While the query bar is focused, the application's global keybindings that conflict with text entry (tab/shift+tab to switch views, `?` to toggle help) SHALL be suppressed so the keystrokes reach the query bar, and SHALL NOT be advertised in the help bar. The quit binding (Ctrl+C) SHALL remain active.

#### Scenario: Tab and help toggle are inert while editing
- **WHEN** the query bar is focused and the user presses `tab` or `?`
- **THEN** the view is not switched and help is not toggled
- **AND** the character (where printable) is entered into the query

#### Scenario: Quit still works while editing
- **WHEN** the query bar is focused and the user presses Ctrl+C
- **THEN** the application quits

### Requirement: Live parsing for validation feedback
While the query bar is being edited, the query SHALL be parsed for validation feedback on Enter and on a debounce of 2 seconds after the last keystroke. Live parsing SHALL update only the error display; it SHALL NOT reload the list.

#### Scenario: Debounced parse updates error display
- **WHEN** the user stops typing for 2 seconds
- **THEN** the query is parsed and the error display is updated (cleared on success, shown on failure)
- **AND** the listed tasks are not reloaded

### Requirement: Apply query on Enter
Pressing Enter SHALL parse the query and, on success, reload the list via ListTasks with the parsed TaskFilter. On parse failure Enter SHALL NOT reload the list.

#### Scenario: Apply a valid query
- **WHEN** the user types `status:done kind:delegated bob` and presses Enter
- **THEN** the list reloads showing done, delegated tasks matching "bob"

#### Scenario: Enter on an invalid query does not reload
- **WHEN** the user presses Enter on a query that fails to parse
- **THEN** the list is not reloaded and the error is shown

### Requirement: Cancel edit on Esc
Pressing Esc while editing SHALL revert the query bar to the last successfully-applied query without reloading.

#### Scenario: Cancel reverts text
- **WHEN** the user edits the query and presses Esc
- **THEN** the query bar reverts to the last applied query
- **AND** the listed tasks are unchanged

### Requirement: Inline parse-error display with range highlight
When parsing fails, the task list SHALL display the parse error inline and SHALL highlight the offending substring in the query bar using the error's range. It SHALL NOT change the currently displayed results.

#### Scenario: Invalid query highlights the bad token
- **WHEN** the query `status:bogus` fails to parse
- **THEN** an error message is shown
- **AND** the `status:bogus` substring is highlighted in the query bar using the error's range
- **AND** the previously listed tasks remain displayed

### Requirement: Create task with "+" or "insert" key
The task list SHALL allow creating a new task by pressing "+" or "insert", which pushes the task edit overlay for a new task.

#### Scenario: Create new task
- **WHEN** user presses "+" or "insert"
- **THEN** a task edit overlay SHALL be pushed for a new task with default values

### Requirement: Assign project with "p" key
The task list SHALL allow assigning a project to the selected task by pressing "p", which pushes the project picker overlay via an injected factory function. The task list SHALL NOT import project packages.

#### Scenario: Assign project to task
- **WHEN** user selects a task
- **AND** presses "p"
- **THEN** the project picker overlay SHALL be pushed for the selected task

#### Scenario: Picker factory not provided
- **WHEN** the task list is constructed without a picker factory
- **THEN** the "p" keybinding SHALL be disabled

#### Scenario: Keybinding disabled when no task selected
- **WHEN** the task list is empty
- **THEN** the "p" keybinding SHALL be disabled
