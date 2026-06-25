## ADDED Requirements

### Requirement: Status picker lists the entity's statuses with the current one preselected
The status picker overlay SHALL present a single-choice selectfield listing the statuses available for the subject entity, with the entity's **current** status preselected (highlighted). The list SHALL include the current status plus every status reachable from it by a valid transition; statuses not reachable from the current status SHALL be omitted. The same picker behavior SHALL be available for both tasks and projects.

#### Scenario: Open task offers done and dropped, open preselected
- **WHEN** the picker opens for an open task
- **THEN** the options SHALL be open, done, dropped
- **AND** open SHALL be the preselected (highlighted) option

#### Scenario: Done task offers only open, done preselected
- **WHEN** the picker opens for a done task
- **THEN** the options SHALL be done and open
- **AND** done SHALL be the preselected option
- **AND** dropped SHALL NOT be offered

#### Scenario: Open project offers someday, done, dropped
- **WHEN** the picker opens for an open project
- **THEN** the options SHALL be open, someday, done, dropped
- **AND** open SHALL be the preselected option

### Requirement: Selecting a different status applies that transition
Choosing a status other than the current one and confirming SHALL apply the matching service transition with an editable timestamp. The timestamp is captured by a `When` field shown in the same overlay, which appears once the selection differs from the current status; `ctrl+s` saves from anywhere. The mapping SHALL be: open→done = Complete, open→dropped = Drop, done/dropped→open = Reopen; for projects additionally open→someday = Park and someday→open = ReopenProject, someday→dropped = Drop. Project Complete and Drop SHALL cascade to the project's tasks (cascade=true).

#### Scenario: Complete an open task via the picker
- **WHEN** the user opens the picker on an open task, arrows to done, and confirms
- **THEN** the task SHALL be completed via CompleteTask with the confirmed timestamp

#### Scenario: Reopen a done task via the picker
- **WHEN** the user opens the picker on a done task, arrows to open, and confirms
- **THEN** the task SHALL be reopened via ReopenTask

#### Scenario: Park a project via the picker
- **WHEN** the user opens the picker on an open project, arrows to someday, and confirms
- **THEN** the project SHALL be parked via ParkProject

### Requirement: Selecting the current status is a no-op
Confirming the picker while the current status is still selected SHALL apply no transition and SHALL dismiss the picker, equivalent to cancelling.

#### Scenario: Confirm without moving off the current status
- **WHEN** the user opens the picker and confirms without changing the selection
- **THEN** no service transition SHALL be called
- **AND** the picker SHALL dismiss leaving the entity unchanged

#### Scenario: Cancel the picker
- **WHEN** the user presses esc in the picker
- **THEN** no transition SHALL be applied and the picker SHALL dismiss
