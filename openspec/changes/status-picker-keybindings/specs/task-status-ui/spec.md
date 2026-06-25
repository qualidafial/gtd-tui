## ADDED Requirements

### Requirement: Change task status with s
Pressing `s` on the selected task SHALL open the status picker overlay seeded with the task's current status and its reachable statuses (open ↔ done/dropped; done/dropped → open). The keymap field carrying this binding SHALL be named `Status` (the key is `s`; its help label SHALL be the fixed string `status`, with no per-selection relabeling). The binding SHALL be enabled whenever a task is selected and disabled when the list is empty.

#### Scenario: Open the status picker on a task
- **WHEN** a task is selected and the user presses `s`
- **THEN** the status picker SHALL be pushed seeded with the task's current status

#### Scenario: Status label is fixed
- **WHEN** the help bar renders the status binding for any task status
- **THEN** the label SHALL read `status` regardless of whether the task is open, done, or dropped

#### Scenario: Status disabled when no selection
- **WHEN** the list is empty
- **THEN** the `s` binding SHALL be disabled and hidden from the help bar

## REMOVED Requirements

### Requirement: Toggle status with space
**Reason**: Replaced by the unified status picker on `s`; the toggle's state-dependent `complete`/`reopen` relabeling is eliminated. `space` is left unbound (reserved for a future peek/preview).
**Migration**: Press `s` to open the status picker, arrow to the target status, and confirm. Complete (open→done) and Reopen (done/dropped→open) are now selections in the picker.
