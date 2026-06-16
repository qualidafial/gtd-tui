## ADDED Requirements

### Requirement: Open task view from project view
Pressing `enter` on a task in the project view SHALL push that task's view screen, mirroring the top-level Tasks tab. The `e` binding SHALL continue to open the task editor. Because the project is already the parent screen, the task view pushed from a project view SHALL disable the go-to-project (`g`) action.

#### Scenario: Enter opens the selected task's view
- **WHEN** the user presses `enter` on a task in the project view
- **THEN** the task view screen is pushed for the selected task

#### Scenario: Edit still opens the editor
- **WHEN** the user presses `e` on a task in the project view
- **THEN** the task edit overlay is pushed for the selected task

#### Scenario: In-project task view disables go-to-project
- **WHEN** a task view is opened from the project view
- **THEN** the go-to-project (`g`) action SHALL NOT be offered, since the project is the parent screen
