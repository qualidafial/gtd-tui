## ADDED Requirements

### Requirement: Change task status with s from task view
Pressing `s` in the task view SHALL push the status picker overlay seeded with the current task's status and its reachable statuses. The picker applies the chosen transition through the existing confirmation overlay.

#### Scenario: Open the status picker from the task view
- **WHEN** the user presses `s` in the task view
- **THEN** the status picker SHALL be pushed seeded with the task's current status

#### Scenario: Complete an open task from the task view
- **WHEN** the user presses `s` in the task view for an open task, arrows to done, and confirms
- **THEN** the task SHALL be completed via CompleteTask

## REMOVED Requirements

### Requirement: Complete or reopen task from task view
**Reason**: Replaced by the unified status picker on `s`. `space` is left unbound (reserved for a future peek/preview).
**Migration**: Press `s` in the task view and select the target status. `delete` is retained as the drop shortcut (see Requirement: Drop task from task view, unchanged).
