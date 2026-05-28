# task-edit-ui Delta Spec

## MODIFIED Requirements

### Requirement: Task editor form fields
The task editor SHALL present a form for the editable fields of a task: Title, Description, Assignee, Due, and Defer Until, in that order. Title SHALL be required (a non-empty validation error blocks saving). Due and Defer Until SHALL use the shared date field (natural-language and `YYYY-MM-DD[ HH:MM]` input). Status SHALL NOT be editable from the editor; status changes are made through the transition confirmation overlay. The Kind field SHALL NOT appear.

#### Scenario: Editor opens on an existing task
- **WHEN** the editor opens for an existing task
- **THEN** the form presents Title, Description, Assignee, Due, and Defer Until

### Requirement: Default values for new tasks
When the editor is opened for a new task (no ID assigned), Status SHALL default to open.

#### Scenario: New task defaults
- **WHEN** the editor opens for a new task
- **THEN** Status is open