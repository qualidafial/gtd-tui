## ADDED Requirements

### Requirement: Bubbletea v2 application structure
The system SHALL use Bubbletea v2 as the terminal UI framework. The root application model SHALL implement the tea.Model interface with Init, Update, and View methods.

#### Scenario: Application starts successfully
- **WHEN** user runs the gtd command
- **THEN** the TUI application initializes and displays the initial view

#### Scenario: Application handles terminal resize
- **WHEN** terminal window is resized
- **THEN** application receives WindowSizeMsg and updates layout accordingly

### Requirement: Entry point in cmd/gtd
The system SHALL provide a main entry point in cmd/gtd/main.go that initializes the Bubbletea program with the root model and runs the event loop.

#### Scenario: Main initializes program
- **WHEN** cmd/gtd/main.go is executed
- **THEN** it creates a tea.Program with the root model
- **AND** runs the program's event loop

#### Scenario: Program exits cleanly
- **WHEN** user quits the application
- **THEN** terminal state is restored to pre-application state

### Requirement: Root model manages application state
The system SHALL have a root model in tui/app.go that manages global application state including the current page, window dimensions, and shared services.

#### Scenario: Root model tracks current page
- **WHEN** application is running
- **THEN** root model knows which page is currently displayed

#### Scenario: Root model tracks window size
- **WHEN** terminal is resized
- **THEN** root model stores current width and height
- **AND** propagates dimensions to active page

### Requirement: Service injection at startup
The system SHALL accept service interfaces (TaskService, ProjectService, InboxService, etc.) at application startup, enabling dependency injection for testing and flexibility.

#### Scenario: Services injected at construction
- **WHEN** application is created
- **THEN** service interfaces are passed to the constructor
- **AND** services are available to all pages

#### Scenario: Mock services for testing
- **WHEN** testing the TUI
- **THEN** mock service implementations can be injected

### Requirement: Quit command exits application
The system SHALL respond to the quit key binding (q or Ctrl+C) by terminating the application cleanly.

#### Scenario: Quit with q key
- **WHEN** user presses q at the top level
- **THEN** application sends tea.Quit command
- **AND** program exits cleanly

#### Scenario: Quit with Ctrl+C
- **WHEN** user presses Ctrl+C
- **THEN** application sends tea.Quit command
- **AND** program exits cleanly
