## MODIFIED Requirements

### Requirement: Task editor form fields
The task editor SHALL present a form for the editable fields of a task: Title, Description, Assignee, Due, and Defer Until, in that order. Title SHALL be required (a non-empty validation error blocks saving). Due and Defer Until SHALL use the shared date field (natural-language and `YYYY-MM-DD[ HH:MM]` input). When editing an existing task, Status SHALL NOT be editable from the editor; status changes for existing tasks are made through the transition confirmation overlay. When creating a new task, the form SHALL offer a terminal Status choice limited to `Open` and `Done` (see "Default values for new tasks"); `Dropped` SHALL NOT be offered at creation. The Kind field SHALL NOT appear.

#### Scenario: Editor opens on an existing task
- **WHEN** the editor opens for an existing task
- **THEN** the form presents Title, Description, Assignee, Due, and Defer Until
- **AND** no Status choice is offered

#### Scenario: Title is required
- **WHEN** the user clears the Title and attempts to submit
- **THEN** a validation error is shown and the task is not saved

### Requirement: Default values for new tasks
When the editor is opened for a new task (no ID assigned), the form SHALL present a terminal inline Status choice offering `Open` and `Done`, defaulting to `Open`. The chosen Status SHALL be the status the task is created in. The default (`Open`) selection SHALL require no extra keypresses relative to the prior single submit affordance.

#### Scenario: New task defaults to open
- **WHEN** the editor opens for a new task
- **THEN** a Status choice offering `Open` and `Done` is shown as the terminal field
- **AND** `Open` is selected by default

#### Scenario: New task offers done but not dropped
- **WHEN** the editor opens for a new task
- **THEN** the Status choice offers exactly `Open` and `Done`
- **AND** `Dropped` is not offered

### Requirement: Save creates or updates
Submitting the form SHALL create the task when it has no ID and update it otherwise. When creating, the task SHALL be created in the Status selected on the form (`Open` or `Done`); a `Done` task is created directly in done status (no order key, `StatusChangedAt` stamped at creation) without an intermediate open state. When the editor was created with a view factory and the submission created a new task, the editor SHALL replace itself with the new task's view via `screen.Replace` (which morphs to the view and batches its window-size request and `Init`), so the parent list reloads when that view is later dismissed. Updates, and creates with no view factory, SHALL dismiss the overlay only (the revealed screen reloads on dismiss).

#### Scenario: Create open on submit
- **WHEN** the form is submitted for a task with no ID and the Status choice is `Open`
- **THEN** the task is created in open status

#### Scenario: Create done on submit
- **WHEN** the form is submitted for a task with no ID and the Status choice is `Done`
- **THEN** the task is created directly in done status
- **AND** the created task has no order key and its StatusChangedAt is set at creation

#### Scenario: Create on submit without view factory
- **WHEN** the form is submitted for a task with no ID and no view factory was supplied
- **THEN** the task is created
- **AND** the overlay is dismissed

#### Scenario: Create on submit with view factory
- **WHEN** the form is submitted for a task with no ID and a view factory was supplied
- **THEN** the task is created
- **AND** the editor is replaced by the new task's view screen

#### Scenario: Update on submit
- **WHEN** the form is submitted for a task with an ID
- **THEN** the task is updated
- **AND** the overlay is dismissed
