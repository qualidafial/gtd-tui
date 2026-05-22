## ADDED Requirements

### Requirement: Page interface contract
The system SHALL define a Page interface that all pages implement. The Page interface SHALL extend tea.Model and provide methods for receiving focus, losing focus, and accessing help bindings.

#### Scenario: Page implements tea.Model
- **WHEN** a page is created
- **THEN** it implements Init, Update, and View methods

#### Scenario: Page receives focus
- **WHEN** page becomes active
- **THEN** page's Focus method is called
- **AND** page can initialize or refresh its state

#### Scenario: Page loses focus
- **WHEN** page becomes inactive
- **THEN** page's Blur method is called
- **AND** page can clean up temporary state

### Requirement: Pages in tui/pages package
The system SHALL organize page implementations under tui/pages/. Each page SHALL be in its own subpackage (e.g., tui/pages/inbox/, tui/pages/tasks/).

#### Scenario: Inbox page location
- **WHEN** looking for the inbox page
- **THEN** it is located at tui/pages/inbox/

#### Scenario: Tasks page location
- **WHEN** looking for the tasks page
- **THEN** it is located at tui/pages/tasks/

### Requirement: Inbox page for unprocessed items
The system SHALL provide an inbox page that displays unprocessed inbox items. The page SHALL support reviewing items one at a time and initiating clarify actions.

#### Scenario: Display inbox items
- **WHEN** inbox page is active
- **THEN** page displays list of unprocessed items

#### Scenario: Select item for clarify
- **WHEN** user selects an inbox item
- **THEN** item details are shown
- **AND** clarify options are available

### Requirement: Tasks page for actionable items
The system SHALL provide a tasks page that displays tasks filterable by status, project, and due date. The page SHALL show next actions by default.

#### Scenario: Display task list
- **WHEN** tasks page is active
- **THEN** page displays list of tasks

#### Scenario: Filter tasks by status
- **WHEN** user applies a status filter
- **THEN** task list shows only matching tasks

#### Scenario: Default shows next actions
- **WHEN** tasks page loads with no filter
- **THEN** page shows next action tasks by default

### Requirement: Projects page for multi-step outcomes
The system SHALL provide a projects page that displays projects filterable by status. The page SHALL show active projects by default.

#### Scenario: Display project list
- **WHEN** projects page is active
- **THEN** page displays list of projects

#### Scenario: Filter projects by status
- **WHEN** user applies a status filter
- **THEN** project list shows only matching projects

#### Scenario: Default shows active projects
- **WHEN** projects page loads with no filter
- **THEN** page shows active projects by default

### Requirement: Someday page for parked ideas
The system SHALL provide a someday page that displays parked ideas sorted by ReviewedAt. The page SHALL surface stalest items first for periodic review.

#### Scenario: Display someday items
- **WHEN** someday page is active
- **THEN** page displays list of someday items

#### Scenario: Sort by stalest first
- **WHEN** someday page loads
- **THEN** items are sorted by ReviewedAt ascending

### Requirement: References page for retrieval content
The system SHALL provide a references page that displays stored reference materials. The page SHALL support search/filter by title.

#### Scenario: Display references
- **WHEN** references page is active
- **THEN** page displays list of reference materials

#### Scenario: Search references
- **WHEN** user enters a search term
- **THEN** references list filters to matching items

### Requirement: Meetings page for meeting records
The system SHALL provide a meetings page that displays meeting records. The page SHALL show upcoming and recent meetings by default.

#### Scenario: Display meetings
- **WHEN** meetings page is active
- **THEN** page displays list of meeting records

#### Scenario: Default shows relevant meetings
- **WHEN** meetings page loads
- **THEN** page shows upcoming and recent meetings

### Requirement: Consistent page styling
The system SHALL apply consistent styling across all pages. Each page SHALL have a header showing the current view name, a content area, and a footer with available actions.

#### Scenario: Page has header
- **WHEN** page is rendered
- **THEN** header shows the page name

#### Scenario: Page has footer
- **WHEN** page is rendered
- **THEN** footer shows available keyboard shortcuts

#### Scenario: Consistent visual style
- **WHEN** navigating between pages
- **THEN** layout structure remains consistent
