## MODIFIED Requirements

### Requirement: Root model manages application state
The system SHALL have a root model in tui/app.go that manages the active screen, window dimensions, and help rendering. The root model SHALL hold a single `active Screen` field instead of a tabs array and overlay slot. It SHALL handle PushMsg, DismissMsg, and InitMsg centrally.

#### Scenario: Root model tracks active screen
- **WHEN** application is running
- **THEN** root model holds the current active Screen (tabContainer or an overlay)

#### Scenario: Root model tracks window size
- **WHEN** terminal is resized
- **THEN** root model stores current width and height
- **AND** propagates dimensions to active screen

#### Scenario: Root model handles push
- **WHEN** root model receives PushMsg
- **THEN** it SHALL wrap the current active in an overlay with the pushed screen as inner
- **AND** it SHALL call Init() on the new active

#### Scenario: Root model handles dismiss
- **WHEN** root model receives DismissMsg
- **AND** the active screen satisfies Popper
- **THEN** it SHALL pop the overlay and call Init() on the restored parent

#### Scenario: Root model handles init
- **WHEN** root model receives InitMsg
- **THEN** it SHALL call Init() on the active screen

### Requirement: Quit command exits application
The system SHALL respond to the quit key binding (Ctrl+C) by terminating the application cleanly. The ? key SHALL toggle help display, suppressed when the active screen is capturing input.

#### Scenario: Quit with Ctrl+C
- **WHEN** user presses Ctrl+C
- **THEN** application sends tea.Quit command
- **AND** program exits cleanly

#### Scenario: Toggle help with ?
- **WHEN** user presses ?
- **AND** the active screen is not capturing input
- **THEN** help display is toggled

#### Scenario: ? suppressed during input capture
- **WHEN** user presses ?
- **AND** CapturingInput(active) returns true
- **THEN** ? is forwarded to the active screen