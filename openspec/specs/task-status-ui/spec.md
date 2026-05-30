# task-status-ui Specification

## Purpose
Defines the task status transition UI: space to toggle, delete to drop, confirmation overlay with editable timestamp, contextual help labels, and reorder constraints.

## Requirements

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

### Requirement: Drop with delete
Pressing `delete` on the selected task SHALL drop it (transition to dropped) only when its status is open, after confirmation. Drop is invalid from a done or dropped task (a done task must be reopened first), so for those statuses `delete` SHALL be a no-op and SHALL NOT be advertised in the help bar.

#### Scenario: Drop an open task
- **WHEN** the selected task is open and the user presses `delete` and confirms
- **THEN** the task is dropped via DropTask and its status becomes dropped

#### Scenario: Delete is inert on a done task
- **WHEN** the selected task is done and the user presses `delete`
- **THEN** nothing happens and no drop confirmation is shown

#### Scenario: Delete is inert on a dropped task
- **WHEN** the selected task is dropped and the user presses `delete`
- **THEN** nothing happens and no drop confirmation is shown

### Requirement: Confirmation overlay for status transitions
Every status transition (Complete, Drop, Reopen) SHALL route through a single shared confirmation overlay before the service call is made. The overlay SHALL present a title, description, and affirmative label appropriate to the target transition, and SHALL preselect the affirmative button so that confirming requires no extra navigation. The overlay SHALL also present an editable transition-timestamp field, prefilled with the current local time, that the user MAY change to record the true time the transition occurred. Confirming SHALL invoke the matching service method (CompleteTask, DropTask, or ReopenTask) with the chosen instant; if the timestamp field is empty the current time SHALL be used. Cancelling SHALL dismiss the overlay without changing the task.

#### Scenario: Confirm applies the transition with the timestamp
- **WHEN** the confirmation overlay is shown for a transition and the user affirms
- **THEN** the corresponding service method is called with the timestamp from the overlay
- **AND** the task list is refreshed

#### Scenario: Affirmative is preselected
- **WHEN** the confirmation overlay first appears
- **THEN** the affirmative button is the default selection, so pressing Enter through the prefilled timestamp immediately confirms

#### Scenario: Timestamp defaults to now
- **WHEN** the confirmation overlay first appears
- **THEN** the transition-timestamp field is prefilled with the current local time

#### Scenario: User backdates the transition
- **WHEN** the user edits the transition-timestamp field to an earlier instant and affirms
- **THEN** the service method is called with that earlier instant

#### Scenario: Empty timestamp falls back to now
- **WHEN** the user clears the transition-timestamp field and affirms
- **THEN** the service method is called with the current time

#### Scenario: Cancel leaves the task unchanged
- **WHEN** the confirmation overlay is shown for a transition and the user cancels
- **THEN** the overlay is dismissed and the task's status is unchanged

### Requirement: Contextual space help label
The help bar SHALL label the `space` binding according to the selected task's status: `complete` when the task is open, and `reopen` when the task is done or dropped.

#### Scenario: Label is "complete" for an open task
- **WHEN** the selected task is open
- **THEN** the help bar shows the `space` binding labeled `complete`

#### Scenario: Label is "reopen" for a closed task
- **WHEN** the selected task is done or dropped
- **THEN** the help bar shows the `space` binding labeled `reopen`

### Requirement: Reorder limited to open tasks
The move bindings (shift+up / shift+down) SHALL be available only when the selected task is open, and only in the direction that keeps the task within the contiguous block of open tasks (which sort above closed ones). Move up SHALL be disabled when the selected task is the first task. Move down SHALL be disabled when the selected task is the last open task (the next task is closed, or there is none). When a binding is disabled it SHALL NOT be advertised in the help bar and pressing it SHALL have no effect.

#### Scenario: Move bindings hidden for a closed task
- **WHEN** the selected task is done or dropped
- **THEN** the help bar does not show the move-up/move-down bindings
- **AND** pressing shift+up or shift+down does not reorder the task

#### Scenario: Move up disabled on the first task
- **WHEN** the selected open task is the first task in the list
- **THEN** the help bar does not show the move-up binding
- **AND** pressing shift+up does not reorder the task

#### Scenario: Move down disabled on the last open task
- **WHEN** the selected open task is the last open task (the next task is closed or absent)
- **THEN** the help bar does not show the move-down binding
- **AND** pressing shift+down does not reorder the task

#### Scenario: Move bindings available for an open task in the interior
- **WHEN** the selected task is open and has open tasks both above and below it
- **THEN** the help bar shows both move-up and move-down bindings
- **AND** pressing them reorders the task

### Requirement: Transitioned task leaves a non-matching filter
After a status transition, a task that no longer matches the active query filter SHALL be removed from the list on refresh. No strike-through or lingering placeholder SHALL be shown.

#### Scenario: Completed task disappears from an open filter
- **WHEN** the active filter is `status:open` and the selected open task is completed
- **THEN** the task is removed from the list after the refresh

### Requirement: Reorder respects active task filter
On the task list, shift+up and shift+down SHALL pass the list's currently active TaskFilter into MoveTaskUp / MoveTaskDown. A single press SHALL produce a one-slot move within the visible (filtered) list rather than within all open tasks. Items outside the filter are not consulted when picking the new position and MAY interleave with filtered items as a result.

#### Scenario: Move down within a search filter takes one press to visibly shift
- **WHEN** the query filter narrows the task list to a subset of open tasks
- **AND** user presses shift+down on a non-last filtered open task
- **THEN** the task SHALL appear one slot lower in the visible list after the reload

#### Scenario: In-project task list move stays within the project
- **WHEN** the task list is rendered inside a project view (using the projectTaskService wrapper)
- **AND** user presses shift+up on a non-first open task
- **THEN** the move SHALL reorder only within that project's open tasks even if the list's TaskFilter has no ProjectID set