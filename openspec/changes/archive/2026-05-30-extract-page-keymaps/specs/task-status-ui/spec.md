## MODIFIED Requirements

### Requirement: Toggle status with space
Pressing `space` on the selected task SHALL initiate a status transition determined by the task's current status: an open task transitions to done (Complete), and a done or dropped task transitions to open (Reopen). The transition SHALL be confirmed before it is applied. The keymap field carrying this binding SHALL be named `ToggleComplete` (the key is `space`; the displayed label flips between `complete` and `reopen` via `SetHelp`).

#### Scenario: Complete an open task
- **WHEN** the selected task is open and the user presses `space` and confirms
- **THEN** the task is completed via CompleteTask and its status becomes done

#### Scenario: Reopen a done task
- **WHEN** the selected task is done and the user presses `space` and confirms
- **THEN** the task is reopened via ReopenTask and its status becomes open

#### Scenario: Reopen a dropped task
- **WHEN** the selected task is dropped and the user presses `space` and confirms
- **THEN** the task is reopened via ReopenTask and its status becomes open

#### Scenario: Binding field name reflects primary action
- **WHEN** code or tests reference the toggle binding via the tasklist keymap
- **THEN** the field SHALL be `tasklist.KeyMap.ToggleComplete`, not `Toggle`