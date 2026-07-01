## ADDED Requirements

### Requirement: Edit due date with `d`
Pressing `d` on the selected task SHALL open the date-editor overlay targeting the task's **Due** date. The overlay SHALL be available wherever a single task is focused: the task list, the task list embedded in the project view, and the task view. The binding SHALL be advertised in the help bar as `d due`.

#### Scenario: Open the due editor from the task list
- **WHEN** a task is selected in the task list and the user presses `d`
- **THEN** the date-editor overlay is pushed targeting the selected task's Due date

#### Scenario: Open the due editor from the task view
- **WHEN** the user presses `d` in the task view
- **THEN** the date-editor overlay is pushed targeting the current task's Due date

#### Scenario: Due binding falls through in the project view
- **WHEN** a task is selected in the project view's embedded task list and the user presses `d`
- **THEN** the date-editor overlay is pushed targeting that task's Due date

### Requirement: Edit defer date with `f`
Pressing `f` on the selected task SHALL open the date-editor overlay targeting the task's **Defer Until** date. The overlay SHALL be available in the same contexts as the `d` binding. The binding SHALL be advertised in the help bar as `f defer`.

#### Scenario: Open the defer editor from the task list
- **WHEN** a task is selected in the task list and the user presses `f`
- **THEN** the date-editor overlay is pushed targeting the selected task's Defer Until date

#### Scenario: Open the defer editor from the task view
- **WHEN** the user presses `f` in the task view
- **THEN** the date-editor overlay is pushed targeting the current task's Defer Until date

### Requirement: Date-editor overlay
The date-editor overlay SHALL present a single date field (built on the shared `datefield` component) plus a confirm button. The field SHALL be prefilled with the task's current value for the target date (Due or Defer Until), or empty when that date is unset. The field label SHALL identify the target date. The overlay SHALL accept natural-language input and echo the resolved absolute date, exactly as `datefield` already does.

#### Scenario: Prefill with the current value
- **WHEN** the overlay opens for a task whose target date is already set
- **THEN** the date field is prefilled with that date

#### Scenario: Empty field for an unset date
- **WHEN** the overlay opens for a task whose target date is unset
- **THEN** the date field is empty

#### Scenario: Natural-language entry
- **WHEN** the user types a natural-language date (e.g. `next friday`) and moves off the field
- **THEN** the field echoes the resolved absolute date that will be saved

### Requirement: Commit sets, changes, or clears the date
Confirming the overlay SHALL persist the entered value to the target date via `TaskService.UpdateTask` and dismiss the overlay. A non-empty parsed date SHALL set that date; an empty field SHALL clear it (set it to nil). Only the target date field SHALL be modified; all other task attributes SHALL be preserved. Pressing `esc` SHALL cancel with no change.

#### Scenario: Set a previously-unset date
- **WHEN** the target date is unset, the user enters a valid date, and confirms
- **THEN** the task's target date is updated to that value via UpdateTask
- **AND** no other task attribute is changed

#### Scenario: Change an existing date
- **WHEN** the target date is already set, the user enters a different valid date, and confirms
- **THEN** the task's target date is updated to the new value via UpdateTask

#### Scenario: Clear a date by submitting empty
- **WHEN** the target date is set, the user clears the field, and confirms
- **THEN** the task's target date becomes nil via UpdateTask

#### Scenario: Cancel leaves the task unchanged
- **WHEN** the overlay is open and the user presses `esc`
- **THEN** the overlay is dismissed and the task is not modified
