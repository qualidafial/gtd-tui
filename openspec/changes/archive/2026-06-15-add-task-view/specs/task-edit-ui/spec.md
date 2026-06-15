## MODIFIED Requirements

### Requirement: Save creates or updates
Submitting the form SHALL create the task when it has no ID and update it otherwise. When the editor was created with a view factory and the submission created a new task, the editor SHALL replace itself with the new task's view via `screen.Replace` (which morphs to the view and batches its window-size request and `Init`), so the parent list reloads when that view is later dismissed. Updates, and creates with no view factory, SHALL dismiss the overlay only (the revealed screen reloads on dismiss).

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