## MODIFIED Requirements

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
</content>
