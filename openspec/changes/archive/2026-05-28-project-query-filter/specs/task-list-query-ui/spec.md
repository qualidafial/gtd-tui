## MODIFIED Requirements

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

### Requirement: Inline parse-error display with range highlight
When parsing fails, the query bar SHALL highlight the offending substring inline using red foreground and underline styling via the shared `querybar` component's `ansi.Cut`-based rendering. The error message SHALL be displayed in the app error bar. The query bar SHALL NOT use a multi-line display for errors.

#### Scenario: Invalid query highlights the bad token
- **WHEN** the query `status:bogus` fails to parse
- **THEN** the `status:bogus` substring is highlighted inline with red foreground and underline
- **AND** the error message is displayed in the app error bar
- **AND** the previously listed tasks remain displayed