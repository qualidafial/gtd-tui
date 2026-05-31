# product-vision Specification

## Purpose
Captures the product intent and guardrails for the GTD TUI: a personal, single-user productivity app for terminal users built around the Getting Things Done methodology, where tasks, projects, and notes live in one cross-linked, navigable interface. This is the authoritative statement of *why the app exists and what it deliberately is not* — the guiding principles and non-goals that constrain every other spec. When a proposed feature conflicts with a principle or non-goal here, this spec governs.

## Requirements

### Requirement: Low ceremony
The app SHALL minimize the overhead of maintaining a GTD system. Capturing a thought, linking it to a project, or finding what to work on next SHALL take seconds, not minutes. Friction in routine operations is treated as a defect, because maintenance overhead is the primary failure mode of GTD systems.

#### Scenario: Capture is fast
- **WHEN** a user captures a thought to the inbox
- **THEN** the operation completes in a single low-friction interaction with no required metadata

### Requirement: One pane of glass
Tasks, projects, and notes SHALL be different views of the same interconnected data, not separate tools. Navigating between them SHALL feel natural rather than like switching apps.

#### Scenario: Navigate between entity types
- **WHEN** a user moves from a task to its project or linked notes
- **THEN** navigation stays within one interface without an app/context switch

### Requirement: Easy linking
Relationships between entities SHALL be first-class. Cross-links SHALL be easy to create and easy to follow. Action items captured in one context (e.g. a meeting note) SHALL surface automatically in the relevant project or inbox.

#### Scenario: Captured action item surfaces
- **WHEN** an action item is captured during a meeting
- **THEN** it appears in the inbox with a link back to its origin

### Requirement: Timeline as context
Every entity SHALL have a chronological history viewable without hunting through separate logs. Because capture is low-friction, timelines SHALL serve as an accurate record of what happened and when, usable for retroactive reports and recalling why a decision was made.

#### Scenario: View entity history
- **WHEN** a user views a task, project, or note
- **THEN** its activity is available in chronological order in context

### Requirement: Personal single-user scope (non-goals)
The app SHALL remain personal, single-user software. The following are explicitly out of scope:
- Team/multi-user collaboration features
- Pomodoro timing or time tracking
- Replacing a calendar. Due dates are supported, but calendar views are out of scope. Calendar scheduling lives in an actual calendar app.

Free-form observation entities (e.g. "hallway conversation," "decision") SHALL be deferred until a concrete use case justifies them; comments-on-edit and meeting notes cover capture-context needs for now.

#### Scenario: Reject out-of-scope feature
- **WHEN** a proposed feature is team collaboration, time tracking, or a calendar view
- **THEN** it is rejected as a non-goal unless this spec is amended first
