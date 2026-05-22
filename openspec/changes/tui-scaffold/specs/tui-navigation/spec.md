## ADDED Requirements

### Requirement: Page-based navigation model
The system SHALL use a page-based navigation model where only one page is active at a time. Pages represent major workflow views (inbox, tasks, projects, etc.).

#### Scenario: Single active page
- **WHEN** application is displaying a page
- **THEN** only that page receives keyboard input
- **AND** only that page is rendered

#### Scenario: Page transition preserves state
- **WHEN** user navigates from page A to page B
- **THEN** page A's state is preserved
- **AND** returning to page A restores previous state

### Requirement: Global navigation keys
The system SHALL provide global navigation keys that work regardless of current page context. These keys SHALL use a consistent modifier pattern to distinguish from page-local keys.

#### Scenario: Navigate to inbox
- **WHEN** user presses the inbox navigation key
- **THEN** application displays the inbox page

#### Scenario: Navigate to tasks
- **WHEN** user presses the tasks navigation key
- **THEN** application displays the tasks page

#### Scenario: Navigate to projects
- **WHEN** user presses the projects navigation key
- **THEN** application displays the projects page

### Requirement: Keyboard-driven navigation
The system SHALL support full keyboard navigation without requiring a mouse. All navigation actions SHALL be accessible via keyboard shortcuts.

#### Scenario: Navigate without mouse
- **WHEN** user operates the application
- **THEN** all navigation is possible using only keyboard

#### Scenario: Discoverable shortcuts
- **WHEN** user views the interface
- **THEN** available keyboard shortcuts are visible or accessible via help

### Requirement: Default page selection
The system SHALL select the default page based on inbox state. If the inbox contains unprocessed items, the inbox page SHALL be shown. Otherwise, the tasks page SHALL be shown.

#### Scenario: Default to inbox when items exist
- **WHEN** application starts with inbox containing items
- **THEN** inbox page is displayed

#### Scenario: Default to tasks when inbox empty
- **WHEN** application starts with empty inbox
- **THEN** tasks page is displayed

### Requirement: Navigation between linked entities
The system SHALL support direct navigation between linked entities. From a task, users SHALL be able to navigate to its project. From a project, users SHALL be able to navigate to its tasks.

#### Scenario: Navigate from task to project
- **WHEN** user selects "go to project" on a task with a project
- **THEN** application displays the project detail view

#### Scenario: Navigate from project to task
- **WHEN** user selects a task within a project view
- **THEN** application displays the task detail view

### Requirement: Navigation history with back
The system SHALL maintain a navigation history stack. Users SHALL be able to go back to the previous view using a back key binding.

#### Scenario: Go back to previous page
- **WHEN** user navigates from tasks to a project detail
- **AND** user presses the back key
- **THEN** application returns to the tasks page

#### Scenario: Back at root is no-op
- **WHEN** user is at the root level of a page
- **AND** user presses the back key
- **THEN** nothing happens (does not exit application)
