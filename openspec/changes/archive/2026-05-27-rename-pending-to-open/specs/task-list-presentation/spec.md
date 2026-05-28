# task-list-presentation Delta Spec

## MODIFIED Requirements

### Requirement: Status marker and title styling
Each task row SHALL begin with a status marker followed by the task title. The marker and title styling SHALL reflect the task status:

- `open` → `[ ]`, default title style.
- `done` → `[x]`, title rendered in a dim "done" color.
- `dropped` → `[-]`, title rendered with strikethrough in a dim "dropped" color.

#### Scenario: Open task marker
- **WHEN** rendering an open task titled "Buy milk"
- **THEN** the row reads `[ ] Buy milk` with the default title style

### Requirement: Assignee chip
A task with a non-nil assignee SHALL display an `@<assignee>` chip.

#### Scenario: Delegated task shows assignee
- **WHEN** rendering a task assigned to "bob"
- **THEN** the row includes the chip `@bob`