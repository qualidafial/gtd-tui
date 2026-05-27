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

#### Scenario: Root model accepts ProjectService
- **WHEN** the application is constructed via tui.New
- **THEN** it SHALL accept a gtd.ProjectService parameter
- **AND** pass it to the project list screen in the tab container

#### Scenario: Tab container includes Projects tab
- **WHEN** the application starts
- **THEN** the tab container SHALL have a "Tasks" tab and a "Projects" tab
- **AND** "Tasks" SHALL be the initially active tab