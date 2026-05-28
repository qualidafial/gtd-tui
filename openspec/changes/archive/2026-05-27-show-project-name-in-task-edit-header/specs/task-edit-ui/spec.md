## MODIFIED Requirements

### Requirement: Non-editable properties header
When editing an existing task (an ID is assigned), the editor SHALL render a compact, dimmed header above the form showing read-only properties: Task ID, Created, Updated, Status, and Project (when linked). Each property SHALL be a label/value line. Created and Updated SHALL render their timestamps in the local timezone. The Project line SHALL display the project's title and SHALL only appear when the task has a non-nil ProjectID. The header SHALL NOT appear when creating a new task.

#### Scenario: Header shown for an existing task with project
- **WHEN** the editor is opened for an existing task that has a ProjectID
- **THEN** a dimmed header shows Task ID, Created, Updated, Status, and Project above the form

#### Scenario: Header shown for an existing task without project
- **WHEN** the editor is opened for an existing task that has no ProjectID
- **THEN** a dimmed header shows Task ID, Created, Updated, and Status above the form
- **AND** no Project line is shown

#### Scenario: No header for a new task
- **WHEN** the editor is opened to create a task
- **THEN** no read-only properties header is shown