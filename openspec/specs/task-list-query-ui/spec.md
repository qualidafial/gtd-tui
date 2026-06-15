# task-list-query-ui Specification

## Purpose
Defines the task list query bar UI: default query, focus/edit behavior, live validation, and apply/cancel semantics.
## Requirements
### Requirement: Query bar on the task list
The task list SHALL display a query bar using the shared `querybar` component, occupying exactly one line above the list. The bar SHALL be seeded with the default query `status:open ready:now` on startup. When the query is empty, the bar SHALL show a placeholder indicating that all tasks are shown.

#### Scenario: Default query on startup
- **WHEN** the task list first loads
- **THEN** the query bar shows `status:open ready:now`
- **AND** only open tasks that are available now (not deferred to the future) are listed

#### Scenario: Empty query shows placeholder
- **WHEN** the query bar is empty
- **THEN** a placeholder indicating all tasks are shown (e.g. `(all tasks)`) is displayed in the bar

#### Scenario: Query bar is single line
- **WHEN** the task list renders
- **THEN** the query bar SHALL occupy exactly one line above the list

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

### Requirement: Live parsing for validation and preview
While the query bar is being edited, the query SHALL be parsed for validation feedback on Enter and on a debounce of 500 milliseconds after the last keystroke. When the debounced parse succeeds, the task list SHALL be reloaded with the new filter as a live preview, but the query bar SHALL remain focused so further edits keep refining the preview. When the debounced parse fails, only the error display SHALL be updated; the list SHALL NOT be reloaded.

#### Scenario: Debounced parse previews the filter
- **WHEN** the user stops typing for 500 milliseconds and the current query parses cleanly
- **THEN** the task list reloads using the parsed filter
- **AND** the query bar remains focused
- **AND** the error display is cleared

#### Scenario: Debounced parse on invalid query updates only the error display
- **WHEN** the user stops typing for 500 milliseconds and the current query fails to parse
- **THEN** the error display is updated to show the parse error
- **AND** the listed tasks are not reloaded

### Requirement: Apply query on Enter
Pressing Enter SHALL parse the query and, on success, reload the list via ListTasks with the parsed TaskFilter. On parse failure Enter SHALL NOT reload the list.

#### Scenario: Apply a valid query
- **WHEN** the user types `status:done assignee:bob` and presses Enter
- **THEN** the list reloads showing done tasks assigned to "bob"

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
When parsing fails, the query bar SHALL highlight the offending substring inline using red foreground and underline styling via the shared `querybar` component's `ansi.Cut`-based rendering. The error message SHALL be displayed in the app error bar. The query bar SHALL NOT use a multi-line display for errors.

#### Scenario: Invalid query highlights the bad token
- **WHEN** the query `status:bogus` fails to parse
- **THEN** the `status:bogus` substring is highlighted inline with red foreground and underline
- **AND** the error message is displayed in the app error bar
- **AND** the previously listed tasks remain displayed

### Requirement: Create task with "+" or "insert" key
The task list SHALL allow creating a new task by pressing "+" or "insert", which pushes the task edit overlay for a new task. The editor SHALL be created with a view factory so that, on successful create, the new task's view screen is shown (the editor dismisses and the task view is pushed).

#### Scenario: Create new task
- **WHEN** user presses "+" or "insert"
- **THEN** a task edit overlay SHALL be pushed for a new task with default values

#### Scenario: New task lands on its view
- **WHEN** the user submits the new-task form successfully
- **THEN** the editor overlay SHALL dismiss
- **AND** the new task's view screen SHALL be pushed

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

### Requirement: Revert to default filter with backslash
The task list SHALL retain its construction-time seed query as the default query.
While the query bar is not focused (the list is focused), pressing `\` SHALL
reset the query bar to the default query, reparse it, and reload the list via
ListTasks with the resulting TaskFilter. The reset SHALL record the default as
the applied query, so that a subsequent Esc reverts to the default rather than to
the previously-typed query. The `\` binding SHALL be disabled — hidden from the
help bar and inert when pressed — whenever the current query already equals the
default query.

#### Scenario: Revert restores the default query
- **WHEN** the user has applied a different query and the query bar is not focused
- **AND** the user presses `\`
- **THEN** the query bar text SHALL revert to the default `status:open ready:now`
- **AND** the list SHALL reload using the default filter
- **AND** the applied query SHALL become the default query

#### Scenario: Esc after revert reverts to the default
- **WHEN** the user has pressed `\` to revert to the default
- **AND** the user later focuses the query bar, edits it, and presses Esc
- **THEN** the query bar SHALL revert to the default query

#### Scenario: Binding disabled when already at default
- **WHEN** the current query equals the default query
- **THEN** the `\` binding SHALL be disabled and SHALL NOT appear in the help bar

#### Scenario: Binding inert while editing
- **WHEN** the query bar is focused for editing
- **AND** the user types `\`
- **THEN** the filter SHALL NOT be reset
- **AND** the `\` character SHALL be entered into the query

#### Scenario: Revert on an empty-default screen clears the filter
- **WHEN** the task list was seeded with an empty default query (all tasks)
- **AND** the user has applied a narrowing query and presses `\`
- **THEN** the query bar SHALL become empty
- **AND** the list SHALL reload showing all tasks

### Requirement: Open task view with enter
Pressing enter on a selected task (while the query bar is not capturing input) SHALL push the task view screen for that task. This replaces the previous behavior where enter opened the task editor directly; the editor is now reached with `e`.

#### Scenario: Enter opens the task view
- **WHEN** the user presses enter on a selected task
- **AND** the query bar is not capturing input
- **THEN** the task view screen SHALL be pushed for the selected task

#### Scenario: Enter ignored while query bar is focused
- **WHEN** the user presses enter while the query bar is focused
- **THEN** the keypress SHALL apply the query and SHALL NOT open the task view

### Requirement: Edit task with "e" key
Pressing `e` on a selected task (while the query bar is not capturing input) SHALL push the task edit overlay for that task. The overlay is created without a view factory, so saving returns to the task list.

#### Scenario: Edit selected task
- **WHEN** the user presses `e` on a selected task
- **THEN** the task edit overlay SHALL be pushed for that task
