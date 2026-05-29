## Why

Meeting notes are a primary source of action items in GTD workflows. Currently there is no way to capture meeting context and spawn linked inbox items from it. This change implements the Meeting entity and MeetingLink infrastructure to enable context-rich action item capture during meetings.

## What Changes

- Add Meeting entity with title, body (markdown), start/end times, and attendees (JSON []string)
- Add MeetingLink join table connecting Meeting to Task, Project, or Item (exactly one FK set, enforced by CHECK constraint)
- Implement MeetingService with CRUD operations and AddActionItem method
- AddActionItem creates inbox Item, links via MeetingLink, and appends line to Meeting body in one transaction
- MeetingLink follows Item through clarification (rewritten to point at resulting entity)
- Add a meetings screen at `tui/pages/meetings/` and register it as a tab in the root tabContainer

## Capabilities

### New Capabilities
- `meeting-crud`: CRUD operations for Meeting entity (create, read, update, delete, list)
- `meeting-action-items`: AddActionItem operation that atomically creates inbox Item, MeetingLink, and updates Meeting body
- `meeting-link-clarification`: MeetingLink rewriting when linked Items are clarified
- `meetings-page`: Meetings screen surfacing upcoming and recent meetings, registered as a tab

### Modified Capabilities

## Impact

- New domain types: Meeting, MeetingLink in root package
- New service interface: MeetingService in root package
- New SQLite tables: meetings, meeting_links with CHECK constraints
- New migration file for schema
- SQLite implementation: sqlite/meeting.go
- Clarify operations must update MeetingLinks when Items are clarified into Tasks or Projects
- **TUI**: New `tui/pages/meetings/` screen wired into `tui.New` as a tab
