## MODIFIED Requirements

### Requirement: Convert to Project wizard
The convert-to-project wizard SHALL be a form-first flow that collects the new project's Title, Outcome, and Description and the re-scoped task's Title and Description, then commits via `ConvertTaskToProject` in a single transaction on submit. Because the source task already exists and is durable, the wizard SHALL NOT use an early checkpoint; abandoning the wizard SHALL leave the task unchanged and standalone. On a successful commit the wizard SHALL open the newly created project's view in place of itself, rather than dismissing to the task list.

#### Scenario: Fields pre-populated from the source task
- **WHEN** the wizard opens for a standalone task
- **THEN** the project Title and Description SHALL be pre-populated from the task (editable in place)
- **AND** the reframed task Title and Description SHALL be pre-populated from the task (editable in place)
- **AND** the project Outcome SHALL start empty

#### Scenario: Commit on submit opens the new project view
- **WHEN** the user submits the wizard
- **THEN** `ConvertTaskToProject` SHALL be called with the collected project and reframed task
- **AND** on success the wizard SHALL be replaced by the view of the newly created project
- **AND** the task list SHALL NOT be the screen the user lands on

#### Scenario: Abandon leaves the task unchanged
- **WHEN** the user cancels the wizard before submitting
- **THEN** no project SHALL be created
- **AND** the task SHALL remain standalone and unchanged

#### Scenario: Commit error displayed
- **WHEN** `ConvertTaskToProject` fails
- **THEN** the error SHALL be displayed in the wizard
- **AND** the wizard SHALL remain open
- **AND** the project view SHALL NOT be opened
