## MODIFIED Requirements

### Requirement: Picker shows open projects and none option
The project picker overlay SHALL display a `selectfield` (the in-house form component) populated with all open projects loaded from ProjectService, plus a `(none)` option (via `WithNone`) to unlink the task from any project. The select SHALL be sized to the overlay's available height.

#### Scenario: Open projects listed
- **WHEN** the picker overlay is opened
- **THEN** all open projects SHALL be listed as options
- **AND** a `(none)` option SHALL appear to allow unlinking

#### Scenario: Current project pre-selected
- **WHEN** the picker is opened for a task that already has a ProjectID
- **THEN** the matching project SHALL be pre-selected in the list via `WithInitialValue`

#### Scenario: No project pre-selects none
- **WHEN** the picker is opened for a task with nil ProjectID
- **THEN** `(none)` SHALL be pre-selected
