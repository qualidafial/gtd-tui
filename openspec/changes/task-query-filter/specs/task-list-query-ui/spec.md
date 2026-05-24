## ADDED Requirements

### Requirement: Query bar on the task list
The task list SHALL display a query bar showing the active query string. The bar SHALL be seeded with the default query `status:pending ready:now` on startup.

#### Scenario: Default query on startup
- **WHEN** the task list first loads
- **THEN** the query bar shows `status:pending ready:now`
- **AND** only pending tasks that are available now (not deferred to the future) are listed

### Requirement: Focus and edit the query
Pressing `/` SHALL focus the query bar for editing. The list's built-in filter keybinding SHALL be disabled so the query bar is the only filtering mechanism.

#### Scenario: Focus query bar
- **WHEN** the user presses `/`
- **THEN** the query bar becomes focused and editable

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
