## Context

The GTD TUI app needs activity timelines to support the Reflect workflow. DESIGN.md specifies that every entity has a history viewable in chronological order, and that low-friction capture makes timelines an accurate historical record.

Current state: The system has entities (Task, Project, Item, Meeting, Comment) with timestamps but no unified event log. Comments exist but are not integrated into a timeline view. Clarification lineage exists (Item.ClarifiedInto) but is not surfaced chronologically.

Constraints:
- Single-user, local SQLite database
- Must integrate with existing transactional service operations
- Timeline queries must be efficient for both entity-scoped and global views

## Goals / Non-Goals

**Goals:**
- Unified timeline entry model capturing all significant events
- Entity-scoped timeline queries (project, task, meeting)
- Global timeline query across all entities
- Timeline entries generated automatically as part of existing operations
- Clarification lineage visible in timelines (Item captured → clarified into Task)

**Non-Goals:**
- Real-time timeline updates (no push notifications, polling is acceptable)
- Timeline entry editing (immutable audit log)
- Aggregation or analytics on timeline data
- Timeline filtering by event type in v1 (can add later)

## Decisions

### Decision: Dedicated timeline_entries table vs. computed view

**Choice**: Dedicated `timeline_entries` table with explicit inserts.

**Rationale**: A computed view (UNION across tables with timestamps) would be complex, slow for global queries, and unable to capture event-specific metadata (e.g., "status changed from X to Y"). A dedicated table allows:
- Efficient indexing for both entity-scoped and time-range queries
- Rich event metadata in a JSON or typed column
- Consistent chronological ordering regardless of source entity

**Alternatives considered**:
- UNION view: Simpler schema but poor query performance and limited metadata
- Event sourcing: Overkill for a single-user app; would require rewriting all services

### Decision: Timeline entry structure

**Choice**: Polymorphic design with entity type/ID columns plus event type and JSON details.

```
timeline_entries:
  id            INTEGER PRIMARY KEY
  entity_type   TEXT NOT NULL  -- 'task', 'project', 'meeting', 'item', 'comment'
  entity_id     INTEGER NOT NULL
  event_type    TEXT NOT NULL  -- 'created', 'updated', 'status_changed', 'clarified', 'commented'
  details       TEXT           -- JSON with event-specific data
  created_at    TEXT NOT NULL  -- ISO8601 UTC
```

**Rationale**: This structure supports:
- Entity-scoped queries: `WHERE entity_type = ? AND entity_id = ?`
- Global queries: `ORDER BY created_at DESC`
- Related entity queries: JSON details can reference parent (e.g., task's project)
- Flexible event metadata without schema migrations for each event type

### Decision: Timeline entry generation location

**Choice**: Service layer generates timeline entries within existing transactions.

**Rationale**: SQLite triggers could auto-generate entries but:
- Cannot capture "before" state for change deltas (status_changed from X to Y)
- Cannot access application context (user intent, related entities)
- Harder to test and debug

Service-layer generation within the same transaction ensures atomicity: if the operation fails, no orphan timeline entry is created.

### Decision: Clarification lineage in timelines

**Choice**: Two timeline entries for clarification - one on the Item, one on the destination entity.

**Rationale**: When Item 5 is clarified into Task 10:
1. Item 5 gets entry: `{event_type: "clarified", details: {into_type: "task", into_id: 10}}`
2. Task 10 gets entry: `{event_type: "created", details: {from_item: 5}}`

This allows viewing the timeline from either perspective and preserves the full lineage chain.

### Decision: Comment representation in timelines

**Choice**: Comments appear as timeline entries with type "commented", linking to the Comment entity.

**Rationale**: Comments are already event-shaped (short text attached to an entity). Rather than duplicating the comment text in the timeline entry, the entry references the comment ID. The timeline query can JOIN to comments for the body text.

## Risks / Trade-offs

**[Risk] Timeline table grows unboundedly** → For a single-user app, growth is limited by human activity. If needed later, add retention policy or archiving. Not a v1 concern.

**[Risk] JSON details column harder to query** → Accept this trade-off. Event metadata varies by type; a typed column per event type would explode the schema. SQLite's `json_extract()` is available if needed.

**[Risk] Service integration increases coupling** → Mitigate by having a single `TimelineService.Record()` method that other services call. Keeps timeline logic in one place.

**[Trade-off] No timeline for Reference entities** → References are static retrieval content without lifecycle changes. Omitting them simplifies the model. Can add later if needed.

**[Trade-off] Someday entities get minimal timeline** → Only creation and promotion events, since Someday has minimal state transitions.

## Open Questions

(none - design is straightforward given the well-defined requirements in DESIGN.md)
