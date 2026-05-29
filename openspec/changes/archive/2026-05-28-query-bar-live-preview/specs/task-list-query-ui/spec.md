## MODIFIED Requirements

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
