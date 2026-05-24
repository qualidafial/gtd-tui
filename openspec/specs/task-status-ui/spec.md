# task-status-ui Specification

## Purpose
TBD - created by syncing change tui-task-status-transitions. Update Purpose after archive.

## Requirements

### Requirement: Toggle status with space
Pressing `space` on the selected task SHALL initiate a status transition determined by the task's current status: a pending task transitions to done (Complete), and a done or dropped task transitions to pending (Reopen). The transition SHALL be confirmed before it is applied.

#### Scenario: Complete a pending task
- **WHEN** the selected task is pending and the user presses `space` and confirms
- **THEN** the task is completed via CompleteTask and its status becomes done

#### Scenario: Reopen a done task
- **WHEN** the selected task is done and the user presses `space` and confirms
- **THEN** the task is reopened via ReopenTask and its status becomes pending

#### Scenario: Reopen a dropped task
- **WHEN** the selected task is dropped and the user presses `space` and confirms
- **THEN** the task is reopened via ReopenTask and its status becomes pending

### Requirement: Drop with delete
Pressing `delete` on the selected task SHALL drop it (transition to dropped) only when its status is pending, after confirmation. Drop is invalid from a done or dropped task (a done task must be reopened first), so for those statuses `delete` SHALL be a no-op and SHALL NOT be advertised in the help bar.

#### Scenario: Drop a pending task
- **WHEN** the selected task is pending and the user presses `delete` and confirms
- **THEN** the task is dropped via DropTask and its status becomes dropped

#### Scenario: Delete is inert on a done task
- **WHEN** the selected task is done and the user presses `delete`
- **THEN** nothing happens and no drop confirmation is shown

#### Scenario: Delete is inert on a dropped task
- **WHEN** the selected task is dropped and the user presses `delete`
- **THEN** nothing happens and no drop confirmation is shown

### Requirement: Confirmation overlay for status transitions
Every status transition (Complete, Drop, Reopen) SHALL route through a single shared confirmation overlay before the service call is made. The overlay SHALL present a title, description, and affirmative label appropriate to the target transition, and SHALL preselect the affirmative button so that confirming requires no extra navigation. Confirming SHALL invoke the matching service method (CompleteTask, DropTask, or ReopenTask); cancelling SHALL dismiss the overlay without changing the task.

#### Scenario: Confirm applies the transition
- **WHEN** the confirmation overlay is shown for a transition and the user affirms
- **THEN** the corresponding service method is called and the task list is refreshed

#### Scenario: Affirmative is preselected
- **WHEN** the confirmation overlay first appears
- **THEN** the affirmative button is the default selection, so pressing Enter immediately confirms

#### Scenario: Cancel leaves the task unchanged
- **WHEN** the confirmation overlay is shown for a transition and the user cancels
- **THEN** the overlay is dismissed and the task's status is unchanged

### Requirement: Contextual space help label
The help bar SHALL label the `space` binding according to the selected task's status: `complete` when the task is pending, and `reopen` when the task is done or dropped.

#### Scenario: Label is "complete" for a pending task
- **WHEN** the selected task is pending
- **THEN** the help bar shows the `space` binding labeled `complete`

#### Scenario: Label is "reopen" for a closed task
- **WHEN** the selected task is done or dropped
- **THEN** the help bar shows the `space` binding labeled `reopen`

### Requirement: Reorder limited to pending tasks
The move bindings (shift+up / shift+down) SHALL be available only when the selected task is pending, and only in the direction that keeps the task within the contiguous block of pending tasks (which sort above closed ones). Move up SHALL be disabled when the selected task is the first task. Move down SHALL be disabled when the selected task is the last pending task (the next task is closed, or there is none). When a binding is disabled it SHALL NOT be advertised in the help bar and pressing it SHALL have no effect.

#### Scenario: Move bindings hidden for a closed task
- **WHEN** the selected task is done or dropped
- **THEN** the help bar does not show the move-up/move-down bindings
- **AND** pressing shift+up or shift+down does not reorder the task

#### Scenario: Move up disabled on the first task
- **WHEN** the selected pending task is the first task in the list
- **THEN** the help bar does not show the move-up binding
- **AND** pressing shift+up does not reorder the task

#### Scenario: Move down disabled on the last pending task
- **WHEN** the selected pending task is the last pending task (the next task is closed or absent)
- **THEN** the help bar does not show the move-down binding
- **AND** pressing shift+down does not reorder the task

#### Scenario: Move bindings available for a pending task in the interior
- **WHEN** the selected task is pending and has pending tasks both above and below it
- **THEN** the help bar shows both move-up and move-down bindings
- **AND** pressing them reorders the task

### Requirement: Transitioned task leaves a non-matching filter
After a status transition, a task that no longer matches the active query filter SHALL be removed from the list on refresh. No strike-through or lingering placeholder SHALL be shown.

#### Scenario: Completed task disappears from a pending filter
- **WHEN** the active filter is `status:pending` and the selected pending task is completed
- **THEN** the task is removed from the list after the refresh
