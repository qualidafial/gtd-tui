## ADDED Requirements

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
