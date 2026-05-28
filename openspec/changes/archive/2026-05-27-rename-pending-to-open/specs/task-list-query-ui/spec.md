# task-list-query-ui Delta Spec

## MODIFIED Requirements

### Requirement: Query bar on the task list
The task list SHALL display a query bar showing the active query string. The bar SHALL be seeded with the default query `status:open ready:now` on startup. When the query is empty, the bar SHALL show a placeholder indicating that all tasks are shown.

#### Scenario: Default query on startup
- **WHEN** the task list first loads
- **THEN** the query bar shows `status:open ready:now`
- **AND** only open tasks that are available now (not deferred to the future) are listed

#### Scenario: Empty query shows placeholder
- **WHEN** the query bar is empty
- **THEN** a placeholder indicating all tasks are shown (e.g. `(all tasks)`) is displayed in the bar