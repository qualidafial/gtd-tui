## MODIFIED Requirements

### Requirement: Focus and edit the query
Pressing `f` SHALL focus the query bar for editing. The list's built-in filter keybinding SHALL be disabled so the query bar is the only filtering mechanism. The `/` key SHALL NOT focus the query bar; it is reserved for a future search feature.

#### Scenario: Focus query bar
- **WHEN** the user presses `f`
- **THEN** the query bar becomes focused and editable

#### Scenario: Slash does not focus the query bar
- **WHEN** the user presses `/` while the list is focused
- **THEN** the query bar SHALL NOT become focused
