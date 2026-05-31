## Why

The Reflect workflow in `openspec/specs/gtd-workflows/spec.md` specifies activity timelines scoped to project, task, or meeting, plus a global timeline across all entities. Because capture is low-friction, timelines become an accurate historical record useful for retroactive reports, status updates, and remembering why decisions were made. This capability is currently unimplemented.

## What Changes

- Add a TimelineEntry domain type representing discrete events (creation, status changes, comments, clarification lineage)
- Add TimelineService interface for querying timeline entries scoped to an entity or globally
- Add SQLite storage for timeline entries with efficient querying by entity or time range
- Integrate timeline generation into existing service operations (Create*, Update*, status transitions, clarify operations)
- Surface timelines in the TUI for the Reflect workflow

## Capabilities

### New Capabilities

- `activity-timelines`: Timeline entry domain model, storage, and query interface for tracking discrete events on entities (creation, status changes, comments, clarification) with both entity-scoped and global timeline views.

### Modified Capabilities

(none - existing specs already mention timelines as a requirement; this implements that requirement)

## Impact

- **Domain layer**: New TimelineEntry type in root package, TimelineService interface
- **SQLite layer**: New timeline_entries table, TimelineService implementation, integration with existing services to emit entries on writes
- **Service layer**: Existing services (Task, Project, Inbox, Meeting, Comment) need to emit timeline entries on state changes
- **TUI layer**: New timeline view component for Reflect workflow
