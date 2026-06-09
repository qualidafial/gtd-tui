## ADDED Requirements

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
