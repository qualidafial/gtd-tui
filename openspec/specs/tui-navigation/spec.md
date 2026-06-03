# tui-navigation Specification

## Purpose
Defines TUI navigation: back-history, tab-based view switching in the tab container, and InputCapturer propagation through containers.

## Requirements

### Requirement: Navigation history with back
The system SHALL maintain a navigation history via overlay parent chains. Users SHALL be able to go back to the previous view using esc. The history is implicit in the overlay structure, not an explicit stack data structure.

#### Scenario: Go back to previous view
- **WHEN** user navigates from a tab to a detail view
- **AND** user presses esc (while not capturing input)
- **THEN** the overlay dismisses and application returns to the previous view

#### Scenario: Back at root is no-op
- **WHEN** user is at the root level (tabContainer)
- **AND** user presses esc
- **THEN** nothing happens (tabContainer does not satisfy Popper)

#### Scenario: Multi-level back
- **WHEN** user has navigated two levels deep (e.g. list -> view -> edit)
- **AND** user presses esc twice
- **THEN** application returns to the original list view
- **AND** each intermediate view reinitializes on restore

### Requirement: Tab-based navigation in tabContainer
The system SHALL provide tab-based navigation within the tabContainer screen. Tab switching SHALL be scoped to tabContainer and naturally suppressed when any overlay is pushed on top.

#### Scenario: Switch tabs with tab key
- **WHEN** tabContainer is the active screen
- **AND** user presses tab
- **THEN** the next tab becomes active
- **AND** Init() is called on the newly active tab

#### Scenario: Tab keys suppressed under overlay
- **WHEN** an overlay is pushed on top of tabContainer
- **AND** user presses tab
- **THEN** the overlay's inner screen receives the key
- **AND** tab switching does not occur

### Requirement: InputCapturer propagation through containers
The system SHALL propagate InputCapturer through container screens (overlay, tabContainer) so that app.Model and overlay wrappers can suppress their keybindings during text input.

#### Scenario: tabContainer delegates to active tab
- **WHEN** the active tab implements InputCapturer and returns true
- **THEN** tabContainer's CapturingInput() SHALL return true

#### Scenario: Overlay delegates to inner screen
- **WHEN** the inner screen implements InputCapturer and returns true
- **THEN** the overlay's CapturingInput() SHALL return true

#### Scenario: Propagation through nested containers
- **WHEN** an overlay wraps a tabContainer whose active tab is capturing input
- **THEN** CapturingInput called on the overlay SHALL return true