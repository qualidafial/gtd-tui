## ADDED Requirements

### Requirement: Change project status with s key
Pressing `s` on the selected project SHALL open the status picker overlay seeded with the project's current status and its reachable statuses (open → someday/done/dropped; someday → open/dropped; done/dropped → open). The picker applies the chosen transition through the existing confirmation overlay, which SHALL continue to show task cascade information for Complete and Drop. The keymap field SHALL be named `Status` with the fixed help label `status`. The binding SHALL be enabled whenever a project is selected.

#### Scenario: Open the status picker on a project
- **WHEN** a project is selected and the user presses `s`
- **THEN** the status picker SHALL be pushed seeded with the project's current status

#### Scenario: Park an open project via the picker
- **WHEN** an open project is selected, the user presses `s`, arrows to someday, and confirms
- **THEN** ParkProject SHALL be called and the list SHALL reload

#### Scenario: Complete an open project shows cascade
- **WHEN** an open project is selected, the user presses `s`, arrows to done, and confirms
- **THEN** the confirmation SHALL show task cascade info and CompleteProject SHALL be called with cascade=true

## RENAMED Requirements

- FROM: `### Requirement: Quick-create project with "+" or "insert" key`
- TO: `### Requirement: Quick-create project with "c" or "insert" key`

## MODIFIED Requirements

### Requirement: Quick-create project with "c" or "insert" key
The system SHALL allow creating a new project by pressing "c" or "insert", which pushes the project edit overlay with an empty project. On submit, it creates an open project with the entered fields.

#### Scenario: Create new project
- **WHEN** user presses "c" or "insert"
- **THEN** the project edit overlay SHALL be pushed with an empty project
- **WHEN** user fills in fields and submits
- **THEN** the system SHALL call CreateProject with status=open and the entered fields
- **AND** the editor overlay SHALL dismiss
- **AND** the project view screen SHALL be pushed for the newly created project

#### Scenario: Cancel project creation
- **WHEN** user presses "c" or "insert" to open the editor
- **AND** user presses escape
- **THEN** the overlay SHALL be dismissed without creating a project

### Requirement: Focus and edit the project query
Pressing `f` SHALL focus the query bar for editing. The list's built-in filter keybinding SHALL be disabled so the query bar is the only filtering mechanism. The `/` key SHALL NOT focus the query bar; it is reserved for a future search feature.

#### Scenario: Focus query bar
- **WHEN** the user presses `f`
- **THEN** the query bar becomes focused and editable

#### Scenario: Slash does not focus the query bar
- **WHEN** the user presses `/` while the list is focused
- **THEN** the query bar SHALL NOT become focused

### Requirement: Keybindings reflect selected project state
The system SHALL enable/disable action keybindings based on the selected project's status. Status change (`s`) SHALL be enabled whenever a project is selected, with the fixed label `status` (no per-status relabeling). The `delete` drop shortcut SHALL be enabled only for open and someday projects. Reorder SHALL remain conditionally enabled by status and position.

#### Scenario: Open project selected
- **WHEN** an open project is selected
- **THEN** `s` (status) SHALL be enabled with label `status`, `delete` SHALL be enabled, and reorder SHALL be conditionally enabled

#### Scenario: Someday project selected
- **WHEN** a someday project is selected
- **THEN** `s` (status) SHALL be enabled, `delete` SHALL be enabled, and reorder SHALL be conditionally enabled

#### Scenario: Done project selected
- **WHEN** a done project is selected
- **THEN** `s` (status) SHALL be enabled, `delete` SHALL be disabled, and reorder SHALL be disabled

#### Scenario: No project selected
- **WHEN** the list is empty
- **THEN** all action bindings except create ("c") SHALL be disabled

## REMOVED Requirements

### Requirement: Toggle project status with space key
**Reason**: Replaced by the unified status picker on `s`; the toggle's `complete`/`reopen` relabeling is eliminated.
**Migration**: Press `s` to open the status picker and select the target status.

### Requirement: Park project with "s" key
**Reason**: Park becomes the `someday` target in the status picker; `s` now opens the picker.
**Migration**: Press `s` and select `someday`.
